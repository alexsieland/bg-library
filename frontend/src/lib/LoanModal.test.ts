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

// Mock apiClient
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

const mockGame = { gameId: "g1", title: "Catan" };
const mockPatronData = {
  patrons: [
    { patronId: "p1", name: "Alice" },
    { patronId: "p2", name: "Bob" },
  ],
};

describe("LoanModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  it("Should search patrons when typing 3 or more characters", async () => {
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name === "Ali") {
        return { patrons: [{ patronId: "p1", name: "Alice" }] };
      }
      return { patrons: [] };
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Ali" } });

    // Wait for debounce (300ms in component)
    await waitFor(
      () => {
        expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: "Ali" });
      },
      { timeout: 1000 },
    );

    await waitFor(() => {
      expect(screen.getByText("Alice")).toBeInTheDocument();
      // Bob should NOT be in the document because we now filter at the backend,
      // and we will mock the return value to ONLY include Alice if name 'Ali' is passed.
    });
  });

  it("Should limit search results to 5 patrons", async () => {
    const manyPatrons = {
      patrons: Array.from({ length: 10 }, (_, i) => ({
        patronId: `${i}`,
        name: `Patron ${i}`,
      })),
    };
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name === "Patron") {
        return manyPatrons;
      }
      return { patrons: [] };
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Patron" } });

    await waitFor(
      () => {
        expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: "Patron" });
        // After it's called, the list should be displayed.
        expect(screen.getByText("Patron 0")).toBeInTheDocument();
      },
      { timeout: 1000 },
    );

    const allButtons = screen.getAllByRole("button", { hidden: true });
    const patronButtons = allButtons.filter((b) =>
      b.textContent?.trim().startsWith("Patron"),
    );
    expect(patronButtons.length).toBe(5);
  });

  it("Should checkout to existing patron when selected", async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatronData);
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    const onLoanSuccess = vi.fn();
    render(LoanModal, { open: true, game: mockGame, onLoanSuccess });

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Ali" } });

    await waitFor(() => screen.getByText("Alice"), { timeout: 3000 });

    const aliceItem = screen.getByText("Alice");
    await fireEvent.click(aliceItem);

    const loanButton = screen.getByText("Loan");
    await fireEvent.click(loanButton);

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p1");
    });
    expect(onLoanSuccess).toHaveBeenCalled();
  });

  it("Should create new patron and checkout if not found", async () => {
    // 1. Initial search (returns nothing matching 'Charlie')
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name === "Charlie") {
        return { patrons: [] };
      }
      return { patrons: [] };
    });
    // 2. Create patron
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p-new",
      name: "Charlie",
    });
    // 3. Checkout
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Charlie" } });

    const loanButton = screen.getByText("Loan");
    await fireEvent.click(loanButton);

    await waitFor(() => {
      // Verify listPatrons call with name
      expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: "Charlie" });
      // Verify Create Patron call
      expect(apiClient.addPatron).toHaveBeenCalledWith({ name: "Charlie" });
      // Verify Checkout call
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p-new");
    });
  });
});

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

    const loanButton = screen.getByText("Loan");
    await fireEvent.click(loanButton);

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith("g1", "p1");
      expect(onLoanSuccess).toHaveBeenCalled();
    });
  });
});
