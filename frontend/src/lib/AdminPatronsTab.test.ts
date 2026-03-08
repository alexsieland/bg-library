import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import AdminPatronsTab from "./AdminPatronsTab.svelte";
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

describe("AdminPatronsTab", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  it("Should render the Patrons heading", () => {
    render(AdminPatronsTab);
    expect(screen.getByText("Patrons")).toBeInTheDocument();
  });

  it("Should render the Add Patron button", () => {
    render(AdminPatronsTab);
    expect(
      screen.getByRole("button", { name: "Add Patron" }),
    ).toBeInTheDocument();
  });

  it("Should not show the Add Patron modal by default", () => {
    render(AdminPatronsTab);
    expect(document.querySelector("dialog")).toBeNull();
  });

  it("Should open the Add Patron modal when the Add Patron button is clicked", async () => {
    render(AdminPatronsTab);
    await fireEvent.click(screen.getByRole("button", { name: "Add Patron" }));
    expect(
      screen.getByPlaceholderText("Enter patron name"),
    ).toBeInTheDocument();
  });

  it("Should show a success toast when a patron is created", async () => {
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p1",
      name: "Alice",
    });

    render(AdminPatronsTab);

    await fireEvent.click(screen.getByRole("button", { name: "Add Patron" }));
    await fireEvent.input(screen.getByPlaceholderText("Enter patron name"), {
      target: { value: "Alice" },
    });
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith(
        "Successfully added patron: Alice",
        "success",
      );
    });
  });

  it("Should close the modal after a patron is successfully created", async () => {
    vi.mocked(apiClient.addPatron).mockResolvedValue({
      patronId: "p1",
      name: "Alice",
    });

    render(AdminPatronsTab);

    await fireEvent.click(screen.getByRole("button", { name: "Add Patron" }));
    await fireEvent.input(screen.getByPlaceholderText("Enter patron name"), {
      target: { value: "Alice" },
    });
    await fireEvent.click(screen.getByTestId("add-patron-submit"));

    // After success the submit button becomes disabled again (modal reset) and toast fires
    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith(
        "Successfully added patron: Alice",
        "success",
      );
      expect(screen.getByTestId("add-patron-submit")).toBeDisabled();
    });
  });

  it("Should re-enable the Add Patron button after the modal is cancelled", async () => {
    render(AdminPatronsTab);

    await fireEvent.click(screen.getByRole("button", { name: "Add Patron" }));
    expect(
      screen.getByPlaceholderText("Enter patron name"),
    ).toBeInTheDocument();

    await fireEvent.click(screen.getByText("Cancel"));

    // Name field should be cleared (modal reset on close)
    await waitFor(() => {
      expect(
        (screen.getByPlaceholderText("Enter patron name") as HTMLInputElement)
          .value,
      ).toBe("");
    });
  });
});
