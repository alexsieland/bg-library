import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import SearchBar from './SearchBar.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('SearchBar', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it('Should update searchQuery when typing', async () => {
    render(SearchBar, { props: { searchQuery: '' } });
    const input = screen.getByRole('searchbox');
    await fireEvent.input(input, { target: { value: 'catan' } });
    expect((input as HTMLInputElement).value).toBe('catan');
  });

  it('Should call onSearch after debounce delay when characters are typed', async () => {
    const onSearch = vi.fn();
    render(SearchBar, { props: { onSearch, searchQuery: '' } });
    const input = screen.getByRole('searchbox');
    
    await fireEvent.input(input, { target: { value: 'cat' } });
    
    // Should not be called immediately
    expect(onSearch).not.toHaveBeenCalled();
    
    // Advance time by 2000ms (default delay)
    vi.advanceTimersByTime(2000);
    
    expect(onSearch).toHaveBeenCalledWith('cat');
  });

  it('Should call onSearch immediately when Enter is pressed', async () => {
    const onSearch = vi.fn();
    render(SearchBar, { props: { onSearch, searchQuery: 'catan' } });
    const input = screen.getByRole('searchbox');
    
    await fireEvent.keyDown(input, { key: 'Enter' });
    
    expect(onSearch).toHaveBeenCalledWith('catan');
  });

  it('Should call onSearch immediately when search button is clicked', async () => {
    const onSearch = vi.fn();
    render(SearchBar, { props: { onSearch, searchQuery: 'catan' } });
    
    // Now we added a button explicitly
    const button = screen.getByRole('button');
    await fireEvent.click(button);
    
    expect(onSearch).toHaveBeenCalledWith('catan');
  });

  it('Should focus the input when a printable character is typed globally', async () => {
    render(SearchBar, { props: { searchQuery: '' } });
    const input = screen.getByRole('searchbox');
    
    expect(document.activeElement).not.toBe(input);
    
    await fireEvent.keyDown(window, { key: 'a' });
    
    expect(document.activeElement).toBe(input);
  });

  it('Should not focus the input when typing globally while already editing another input', async () => {
    render(SearchBar, { props: { searchQuery: '' } });
    const input = screen.getByRole('searchbox');
    
    // Create another input and focus it
    const otherInput = document.createElement('input');
    document.body.appendChild(otherInput);
    otherInput.focus();
    expect(document.activeElement).toBe(otherInput);
    
    await fireEvent.keyDown(window, { key: 'b' });
    
    expect(document.activeElement).toBe(otherInput);
    expect(document.activeElement).not.toBe(input);
    
    document.body.removeChild(otherInput);
  });
});
