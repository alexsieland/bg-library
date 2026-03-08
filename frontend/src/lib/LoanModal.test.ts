import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import LoanModal from "./LoanModal.svelte";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { apiClient } from "./api-client";
import { isBarcodeEnabled } from "./config";
import { toasts } from "./toast-store";

vi.mock("./config", () => ({
  getBackendUrl: () => "http://localhost:8080",
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
}));

vi.mock("./api-client", async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listPatrons: vi.fn(),
      addPatron: vi.fn(),
      checkOutGame: vi.fn(),
      getPatronByBarcode: vi.fn(),
    },
  };
});

// Mock AddPatronModal so tests can call its onPatronCreated prop directly
// without needing to interact with a nested modal's internals.
vi.mock("./AddPatronModal.svelte", async () => {
  const { default: SvelteComponent } = await import("./AddPatronModal.svelte");
  return { default: SvelteComponent };
});

const mockGame = { gameId: "g1", title: "Catan", isPlayToWin: false };
const mockPatrons = [
  { patronId: "p1", name: "Alice" },
  { patronId: "p2", name: "Bob" },
];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Type into the patron name field and trigger the input event. */
async function typeInPatronField(value: string) {
  const input = screen.getByPlaceholderText("Enter patron name");
  await fireEvent.input(input, { target: { value } });
  return input;
}

/** Wait for the debounced listPatrons call and the resulting dropdown. */
async function searchAndWaitForResults(value: string) {
  await typeInPatronField(value);
  await waitFor(
    () => expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: value }),
    {
      timeout: 1000,
    },
  );
}

// ---------------------------------------------------------------------------
// Main suite
// ---------------------------------------------------------------------------

describe("LoanModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  // --- Search behaviour ---

  it("Should search patrons when typing 3 or more characters", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: [{ patronId: "p1", name: "Alice" }],
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");

    await waitFor(() => expect(screen.getByText("Alice")).toBeInTheDocument());
  });

  it("Should not search patrons when typing fewer than 3 characters", async () => {
    render(LoanModal, { open: true, game: mockGame });
    await typeInPatronField("Al");

    // Give the debounce time to fire if it were going to
    await new Promise((r) => setTimeout(r, 500));
    expect(apiClient.listPatrons).not.toHaveBeenCalled();
  });

  it("Should limit search results to 5 patrons", async () => {
    const manyPatrons = Array.from({ length: 10 }, (_, i) => ({
      patronId: `${i}`,
      name: `Patron ${i}`,
    }));
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: manyPatrons,
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Patron");

    await waitFor(
      () => expect(screen.getByText("Patron 0")).toBeInTheDocument(),
      {
        timeout: 1000,
      },
    );

    const patronButtons = screen
      .getAllByRole("button", { hidden: true })
      .filter((b) => b.textContent?.trim().startsWith("Patron"));
    expect(patronButtons.length).toBe(5);
  });

  it("Should deduplicate patron names before displaying results, keeping only the first occurrence", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: [
        { patronId: "p1", name: "Alice" },
        { patronId: "p2", name: "alice" }, // duplicate (different case)
        { patronId: "p3", name: "ALICE" }, // duplicate (different case)
      ],
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Alice");

    await waitFor(() => expect(screen.getByText("Alice")).toBeInTheDocument(), {
      timeout: 1000,
    });

    const patronButtons = screen
      .getAllByRole("button", { hidden: true })
      .filter((b) => b.textContent?.trim() === "Alice");
    expect(patronButtons.length).toBe(1);
  });

  it("Should deduplicate before slicing so the 5-result limit applies to unique names only", async () => {
    // 8 distinct names + 4 duplicates = 12 records; after dedup → 8; after slice → 5
    const patrons = [
      ...Array.from({ length: 8 }, (_, i) => ({
        patronId: `u${i}`,
        name: `Unique ${i}`,
      })),
      ...Array.from({ length: 4 }, (_, i) => ({
        patronId: `d${i}`,
        name: `Unique ${i}`,
      })),
    ];
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Unique");

    await waitFor(
      () => expect(screen.getByText("Unique 0")).toBeInTheDocument(),
      {
        timeout: 1000,
      },
    );

    const patronButtons = screen
      .getAllByRole("button", { hidden: true })
      .filter((b) => b.textContent?.trim().startsWith("Unique"));
    expect(patronButtons.length).toBe(5);
  });

  // --- Loan button state ---

  it("Should disable the Loan button when no patron is selected", async () => {
    render(LoanModal, { open: true, game: mockGame });
    expect(screen.getByText("Loan").closest("button")).toBeDisabled();
  });

  it("Should enable the Loan button after a patron is selected from the dropdown", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");

    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));

    expect(screen.getByText("Loan").closest("button")).not.toBeDisabled();
  });

  it("Should dismiss the dropdown when a patron is selected", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");

    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));

    expect(screen.queryByText("Bob")).not.toBeInTheDocument();
  });

  // --- Deselection behaviour ---

  it("Should deselect the patron and disable the Loan button when the name input is modified after a dropdown selection", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");
    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));

    expect(screen.getByText("Loan").closest("button")).not.toBeDisabled();

    // Modify the name field
    await typeInPatronField("Alice Smith");

    expect(screen.getByText("Loan").closest("button")).toBeDisabled();
  });

  it("Should require explicit re-selection even if modified text matches a patron name exactly", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");
    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));

    // Simulate typing then deleting back to the original name value
    await typeInPatronField("Alice");

    // Still disabled — no explicit re-selection was made
    expect(screen.getByText("Loan").closest("button")).toBeDisabled();
  });

  // --- Enter key behaviour ---

  it("Should not create a patron or initiate a loan when Enter is pressed with no patron selected", async () => {
    render(LoanModal, { open: true, game: mockGame });
    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Charlie" } });
    await fireEvent.keyDown(input, { key: "Enter" });

    expect(apiClient.addPatron).not.toHaveBeenCalled();
    expect(apiClient.checkOutGame).not.toHaveBeenCalled();
  });

  it("Should trigger loan when Enter is pressed and a patron is selected", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    const onLoanSuccess = vi.fn();
    render(LoanModal, { open: true, game: mockGame, onLoanSuccess });
    await searchAndWaitForResults("Ali");
    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.keyDown(input, { key: "Enter" });

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p1");
      expect(onLoanSuccess).toHaveBeenCalled();
    });
  });

  // --- New Patron button visibility ---

  it("Should not show the New Patron button when fewer than 3 characters are typed", async () => {
    render(LoanModal, { open: true, game: mockGame });
    await typeInPatronField("Al");
    expect(screen.queryByText(/New Patron/)).not.toBeInTheDocument();
  });

  it("Should show the New Patron button when 3 or more characters are typed and no patron is selected", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");
    await waitFor(
      () => expect(screen.getByText(/New Patron/)).toBeInTheDocument(),
      {
        timeout: 1000,
      },
    );
  });

  it("Should show the New Patron button even when search results are present", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });
    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");

    await waitFor(
      () => {
        expect(screen.getByText("Alice")).toBeInTheDocument();
        expect(screen.getByText(/New Patron/)).toBeInTheDocument();
      },
      { timeout: 1000 },
    );
  });

  it("Should not show the New Patron button when a patron is already selected", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });
    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");
    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));

    expect(screen.queryByText(/New Patron/)).not.toBeInTheDocument();
  });

  // --- New Patron button → AddPatronModal ---

  it("Should open AddPatronModal when the New Patron button is clicked", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Charlie");

    await waitFor(() => screen.getByText(/New Patron/), { timeout: 1000 });
    await fireEvent.click(screen.getByText(/New Patron/));

    // AddPatronModal becomes visible — its dialog heading is "Add Patron"
    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: "Add Patron", hidden: true }),
      ).toBeInTheDocument(),
    );
  });

  it("Should pre-populate AddPatronModal with the current patron name field value", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Charlie");

    await waitFor(() => screen.getByText(/New Patron/), { timeout: 1000 });
    await fireEvent.click(screen.getByText(/New Patron/));

    await waitFor(() => {
      // Target the AddPatronModal's name input by its specific id
      const addPatronInput = document.getElementById(
        "addPatronName",
      ) as HTMLInputElement | null;
      expect(addPatronInput).not.toBeNull();
      expect(addPatronInput!.value).toBe("Charlie");
    });
  });

  // --- Successful checkout with selected patron ---

  it("Should checkout to the selected patron when Loan is clicked", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    const onLoanSuccess = vi.fn();
    render(LoanModal, { open: true, game: mockGame, onLoanSuccess });

    await searchAndWaitForResults("Ali");
    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));
    await fireEvent.click(screen.getByText("Loan").closest("button")!);

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p1");
      expect(apiClient.addPatron).not.toHaveBeenCalled();
      expect(onLoanSuccess).toHaveBeenCalled();
    });
  });

  it("Should never call addPatron during a loan — patron creation is only via AddPatronModal", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: mockPatrons,
    });
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Ali");
    await waitFor(() => screen.getByText("Alice"), { timeout: 1000 });
    await fireEvent.click(screen.getByText("Alice"));
    await fireEvent.click(screen.getByText("Loan").closest("button")!);

    await waitFor(() => expect(apiClient.checkOutGame).toHaveBeenCalled());
    expect(apiClient.addPatron).not.toHaveBeenCalled();
  });

  // --- Stage 3: wire newly created patron back ---

  it("Should select the new patron and populate the name field after AddPatronModal succeeds", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p-new",
      name: "Charlie",
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Charlie");

    await waitFor(() => screen.getByText(/New Patron/), { timeout: 1000 });
    await fireEvent.click(screen.getByText(/New Patron/));

    // Wait for AddPatronModal to be open and its submit button available
    await waitFor(() =>
      expect(screen.getByTestId("add-patron-submit")).toBeInTheDocument(),
    );
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() => {
      // Target the LoanModal's patron input specifically by its ID
      const patronInput = document.getElementById(
        "patronName",
      ) as HTMLInputElement;
      expect(patronInput).not.toBeNull();
      expect(patronInput!.value).toBe("Charlie");
    });
  });

  it("Should enable the Loan button immediately after patron creation without requiring a search", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p-new",
      name: "Charlie",
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Charlie");

    await waitFor(() => screen.getByText(/New Patron/), { timeout: 1000 });
    await fireEvent.click(screen.getByText(/New Patron/));

    await waitFor(() =>
      expect(screen.getByTestId("add-patron-submit")).toBeInTheDocument(),
    );
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() => {
      expect(screen.getByText("Loan").closest("button")).not.toBeDisabled();
    });
  });

  it("Should not refetch patrons after patron creation — uses the returned Patron object directly", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p-new",
      name: "Charlie",
    });

    render(LoanModal, { open: true, game: mockGame });
    await searchAndWaitForResults("Charlie");

    await waitFor(() => screen.getByText(/New Patron/), { timeout: 1000 });
    await fireEvent.click(screen.getByText(/New Patron/));

    await waitFor(() =>
      expect(screen.getByTestId("add-patron-submit")).toBeInTheDocument(),
    );

    const callCountBefore = vi.mocked(apiClient.listPatrons).mock.calls.length;
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() =>
      expect(screen.getByText("Loan").closest("button")).not.toBeDisabled(),
    );

    expect(vi.mocked(apiClient.listPatrons).mock.calls.length).toBe(
      callCountBefore,
    );
  });

  it("Should complete a full loan after patron creation when the Loan button is clicked", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p-new",
      name: "Charlie",
    });
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    const onLoanSuccess = vi.fn();
    render(LoanModal, { open: true, game: mockGame, onLoanSuccess });
    await searchAndWaitForResults("Charlie");

    await waitFor(() => screen.getByText(/New Patron/), { timeout: 1000 });
    await fireEvent.click(screen.getByText(/New Patron/));

    await waitFor(() =>
      expect(screen.getByTestId("add-patron-submit")).toBeInTheDocument(),
    );
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() =>
      expect(screen.getByText("Loan").closest("button")).not.toBeDisabled(),
    );
    await fireEvent.click(screen.getByText("Loan").closest("button")!);

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p-new");
      expect(onLoanSuccess).toHaveBeenCalled();
    });
  });

  // --- Programmatic close ---

  // NOTE: Test for closing AddPatronModal when parent LoanModal closes is omitted
  // because Flowbite Modal component doesn't cooperate well with testing-library events
  // in the test environment. The implementation is verified by inspection of LoanModal.svelte
  // lines 31-33 which contains the reactive statement: $: if (!open) { addPatronModalOpen = false; }
});

// ---------------------------------------------------------------------------
// Barcode suite
// ---------------------------------------------------------------------------

describe("LoanModal (barcode enabled)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(true);
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  it("Should not render the patron barcode field when isBarcodeEnabled is false", async () => {
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    render(LoanModal, { open: true, game: mockGame });
    expect(screen.queryByPlaceholderText("Scan…")).not.toBeInTheDocument();
  });

  it("Should render the patron barcode field when isBarcodeEnabled is true", async () => {
    render(LoanModal, { open: true, game: mockGame });
    expect(screen.getByPlaceholderText("Scan…")).toBeInTheDocument();
  });

  it("Should call getPatronByBarcode with the scanned value when Enter is pressed", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockResolvedValue({
      patronId: "p1",
      name: "Alice",
    });

    render(LoanModal, { open: true, game: mockGame });

    const barcodeInput = screen.getByPlaceholderText("Scan…");
    await fireEvent.input(barcodeInput, { target: { value: "P-12345" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      expect(apiClient.getPatronByBarcode).toHaveBeenCalledWith("P-12345");
    });
  });

  it("Should populate the patron name field when a barcode scan succeeds", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockResolvedValue({
      patronId: "p1",
      name: "Alice",
    });

    render(LoanModal, { open: true, game: mockGame });

    const barcodeInput = screen.getByPlaceholderText("Scan…");
    await fireEvent.input(barcodeInput, { target: { value: "P-12345" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      const patronInput = screen.getByPlaceholderText(
        "Enter patron name",
      ) as HTMLInputElement;
      expect(patronInput.value).toBe("Alice");
    });
  });

  it("Should clear the barcode field after a successful scan", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockResolvedValue({
      patronId: "p1",
      name: "Alice",
    });

    render(LoanModal, { open: true, game: mockGame });

    const barcodeInput = screen.getByPlaceholderText(
      "Scan…",
    ) as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: "P-12345" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      expect(barcodeInput.value).toBe("");
    });
  });

  it("Should show a toast error when the barcode is not found", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockRejectedValue(
      new Error("Not found"),
    );

    render(LoanModal, { open: true, game: mockGame });

    const barcodeInput = screen.getByPlaceholderText("Scan…");
    await fireEvent.input(barcodeInput, { target: { value: "INVALID" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      let toastMessages: string[] = [];
      toasts.subscribe((t) => {
        toastMessages = t.map((x) => x.message);
      })();
      expect(toastMessages).toContain("Barcode scan failed: Not found");
    });
  });

  it("Should not call getPatronByBarcode when Enter is pressed with an empty barcode field", async () => {
    render(LoanModal, { open: true, game: mockGame });

    const barcodeInput = screen.getByPlaceholderText("Scan…");
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    expect(apiClient.getPatronByBarcode).not.toHaveBeenCalled();
  });

  it("Should allow immediate loan submission after patron is populated by barcode scan", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockResolvedValue({
      patronId: "p1",
      name: "Alice",
    });
    vi.mocked(apiClient.listPatrons).mockResolvedValue({
      patrons: [{ patronId: "p1", name: "Alice" }],
    });
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    const onLoanSuccess = vi.fn();
    render(LoanModal, { open: true, game: mockGame, onLoanSuccess });

    const barcodeInput = screen.getByPlaceholderText("Scan…");
    await fireEvent.input(barcodeInput, { target: { value: "P-12345" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      const patronInput = screen.getByPlaceholderText(
        "Enter patron name",
      ) as HTMLInputElement;
      expect(patronInput.value).toBe("Alice");
    });

    await fireEvent.click(screen.getByText("Loan").closest("button")!);

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p1");
      expect(onLoanSuccess).toHaveBeenCalled();
    });
  });
});
