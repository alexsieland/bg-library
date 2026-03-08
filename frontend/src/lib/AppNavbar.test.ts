import { render, screen } from "@testing-library/svelte";
import AppNavbar from "./AppNavbar.svelte";
import { describe, it, expect } from "vitest";

describe("AppNavbar", () => {
  it('Should display the brand name "BG Library"', () => {
    render(AppNavbar);
    expect(screen.getByText("BG Library")).toBeInTheDocument();
  });

  it('Should highlight the "Check Out" tab when activeTab is "checkout"', () => {
    render(AppNavbar, { props: { activeTab: "checkout" } });
    const checkOutLink = screen.getByText("Check Out");
    expect(checkOutLink).toHaveClass("text-blue-300");
  });

  it('Should highlight the "Check In" tab when activeTab is "checkin"', () => {
    render(AppNavbar, { props: { activeTab: "checkin" } });
    const checkInLink = screen.getByText("Check In");
    expect(checkInLink).toHaveClass("text-blue-300");
  });

  it('Should highlight the "Admin" tab when activeTab is "admin"', () => {
    render(AppNavbar, { props: { activeTab: "admin" } });
    const adminLink = screen.getByText("Admin");
    expect(adminLink).toHaveClass("text-blue-300");
  });
});
