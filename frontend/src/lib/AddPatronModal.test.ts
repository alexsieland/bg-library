import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import AddPatronModal from "./AddPatronModal.svelte";
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
      addPatron: vi.fn(),
      getPatronByBarcode: vi.fn(),
    },
  };
});

vi.mock("./toast-store", () => ({
  toasts: {
    add: vi.fn(),
  },
}));

const mockPatron = { patronId: "p1", name: "Alice" };

describe("AddPatronModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  it("Should render the patron name input when open", async () => {
    render(AddPatronModal, { open: true });
    expect(
      screen.getByPlaceholderText("Enter patron name"),
    ).toBeInTheDocument();
  });

  it("Should pre-populate the name field with initialName when opened", async () => {
    render(AddPatronModal, { open: true, initialName: "Bob" });
    const input = screen.getByPlaceholderText(
      "Enter patron name",
    ) as HTMLInputElement;
    expect(input.value).toBe("Bob");
  });

  it("Should disable the Add Patron button when the name field is empty", async () => {
    render(AddPatronModal, { open: true });
    expect(screen.getByTestId("add-patron-submit")).toBeDisabled();
  });

  it("Should enable the Add Patron button when the name field has content", async () => {
    render(AddPatronModal, { open: true });
    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Alice" } });
    expect(screen.getByTestId("add-patron-submit")).not.toBeDisabled();
  });

  it("Should not submit when Enter is pressed in the name input", async () => {
    render(AddPatronModal, { open: true });
    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Alice" } });
    await fireEvent.keyDown(input, { key: "Enter" });
    expect(apiClient.addPatron).not.toHaveBeenCalled();
  });

  it("Should call onPatronCreated with the new patron on successful submit", async () => {
    vi.mocked(apiClient.addPatron).mockResolvedValue(mockPatron);
    const onPatronCreated = vi.fn();

    render(AddPatronModal, { open: true, onPatronCreated });

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Alice" } });
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() => {
      expect(apiClient.addPatron).toHaveBeenCalledWith({ name: "Alice" });
      expect(onPatronCreated).toHaveBeenCalledWith(mockPatron);
    });
  });

  it("Should call onCancel when the Cancel button is clicked", async () => {
    const onCancel = vi.fn();
    render(AddPatronModal, { open: true, onCancel });
    await fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalled();
  });

  it("Should show an error toast when addPatron fails", async () => {
    vi.mocked(apiClient.addPatron).mockRejectedValue(new Error("Server error"));

    render(AddPatronModal, { open: true });

    const input = screen.getByPlaceholderText("Enter patron name");
    await fireEvent.input(input, { target: { value: "Alice" } });
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith(
        "Failed to add patron: Server error",
        "error",
      );
    });
  });

  it("Should reset fields when the modal is closed after cancellation", async () => {
    const { rerender } = render(AddPatronModal, {
      open: true,
      initialName: "Alice",
    });

    expect(
      (screen.getByPlaceholderText("Enter patron name") as HTMLInputElement)
        .value,
    ).toBe("Alice");

    await fireEvent.click(screen.getByText("Cancel"));

    await rerender({ open: true, initialName: "" });

    await waitFor(() => {
      expect(
        (screen.getByPlaceholderText("Enter patron name") as HTMLInputElement)
          .value,
      ).toBe("");
    });
  });

  it("Should not show the barcode input when isBarcodeEnabled is false", async () => {
    render(AddPatronModal, { open: true });
    expect(
      screen.queryByPlaceholderText("Scan patron barcode…"),
    ).not.toBeInTheDocument();
  });
});

describe("AddPatronModal (barcode enabled)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(true);
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  it("Should show the barcode input when isBarcodeEnabled is true", async () => {
    render(AddPatronModal, { open: true });
    expect(
      screen.getByPlaceholderText("Scan patron barcode…"),
    ).toBeInTheDocument();
  });

  it("Should not submit when Enter is pressed in the barcode input", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockRejectedValue(
      new Error("Not found"),
    );

    render(AddPatronModal, { open: true });

    const barcodeInput = screen.getByPlaceholderText("Scan patron barcode…");
    await fireEvent.input(barcodeInput, { target: { value: "1234567890" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    expect(apiClient.addPatron).not.toHaveBeenCalled();
  });

  it("Should show an error toast and clear the barcode field when the barcode already belongs to a patron", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockResolvedValue({
      patronId: "p2",
      name: "Bob",
    });

    render(AddPatronModal, { open: true });

    const barcodeInput = screen.getByPlaceholderText(
      "Scan patron barcode…",
    ) as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: "9780307455925" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith(
        "A patron with this barcode already exists",
        "error",
      );
      expect(barcodeInput.value).toBe("");
    });
  });

  it("Should keep the barcode in the field when the barcode is not found (free to use)", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockRejectedValue(
      new Error("Not found"),
    );

    render(AddPatronModal, { open: true });

    const barcodeInput = screen.getByPlaceholderText(
      "Scan patron barcode…",
    ) as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: "1234567890" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });

    await waitFor(() => {
      expect(barcodeInput.value).toBe("1234567890");
      expect(toasts.add).not.toHaveBeenCalled();
    });
  });

  it("Should submit with the barcode value included when a patron is created", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockRejectedValue(
      new Error("Not found"),
    );
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      ...mockPatron,
      barcode: "1234567890",
    });
    const onPatronCreated = vi.fn();

    render(AddPatronModal, { open: true, onPatronCreated });

    await fireEvent.input(screen.getByPlaceholderText("Enter patron name"), {
      target: { value: "Alice" },
    });

    const barcodeInput = screen.getByPlaceholderText(
      "Scan patron barcode…",
    ) as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: "1234567890" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });
    await waitFor(() => expect(barcodeInput.value).toBe("1234567890"));

    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() => {
      expect(apiClient.addPatron).toHaveBeenCalledWith({
        name: "Alice",
        barcode: "1234567890",
      });
      expect(onPatronCreated).toHaveBeenCalled();
    });
  });

  it("Should reset name and barcode fields when the modal is closed after success", async () => {
    vi.mocked(apiClient.getPatronByBarcode).mockRejectedValue(
      new Error("Not found"),
    );
    vi.mocked(apiClient.addPatron).mockResolvedValue(mockPatron);

    const { rerender } = render(AddPatronModal, {
      open: true,
      initialName: "Alice",
    });

    const barcodeInput = screen.getByPlaceholderText(
      "Scan patron barcode…",
    ) as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: "1234567890" } });
    await fireEvent.keyDown(barcodeInput, { key: "Enter" });
    await waitFor(() => expect(barcodeInput.value).toBe("1234567890"));

    await fireEvent.click(screen.getByTestId("add-patron-submit"));
    await waitFor(() => expect(apiClient.addPatron).toHaveBeenCalled());

    await rerender({ open: true, initialName: "" });

    await waitFor(() => {
      expect(
        (screen.getByPlaceholderText("Enter patron name") as HTMLInputElement)
          .value,
      ).toBe("");
      expect(
        (
          screen.getByPlaceholderText(
            "Scan patron barcode…",
          ) as HTMLInputElement
        ).value,
      ).toBe("");
    });
  });
});
