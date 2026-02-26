import { render, cleanup } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import Debounce from './debounce.svelte';
import { tick } from 'svelte';

describe('Debounce Snippet', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });

  it('Should trigger after the specified delay', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: '' };
    
    render(Debounce, {
      value: 'test',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0
    });

    // Initial reactive block runs on mount
    expect(onTrigger).not.toHaveBeenCalled();

    vi.advanceTimersByTime(500);
    expect(onTrigger).toHaveBeenCalledWith('test');
    expect(lastValueRef.v).toBe('test');
  });

  it('Should not trigger if value hasn\'t changed from lastValueRef.v', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: 'test' };
    
    render(Debounce, {
      value: 'test',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0
    });

    vi.advanceTimersByTime(500);
    expect(onTrigger).not.toHaveBeenCalled();
  });

  it('Should not trigger if value length is less than minLength but greater than 0', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: '' };
    
    render(Debounce, {
      value: 'ab',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0,
      minLength: 3
    });

    vi.advanceTimersByTime(500);
    expect(onTrigger).not.toHaveBeenCalled();
  });

  it('Should trigger for empty string even if minLength is set', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: 'previous' };
    
    render(Debounce, {
      value: '',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0,
      minLength: 3
    });

    vi.advanceTimersByTime(500);
    expect(onTrigger).toHaveBeenCalledWith('');
    expect(lastValueRef.v).toBe('');
  });

  it('Should cancel pending timer when cancelKey changes', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: '' };
    
    const { rerender } = render(Debounce, {
      value: 'test',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0
    });

    vi.advanceTimersByTime(250);
    
    // Update cancelKey
    await rerender({
      value: 'test',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 1
    });
    
    // The previous timer was at 250ms.
    // Changing cancelKey cleared it and started a NEW 500ms timer.
    // So if we advance by another 250ms (total 500ms from start), 
    // the NEW timer should still have 250ms left.
    
    vi.advanceTimersByTime(250);
    expect(onTrigger).not.toHaveBeenCalled();
    
    vi.advanceTimersByTime(250); // total 500 since cancelKey change
    expect(onTrigger).toHaveBeenCalledWith('test');
  });

  it('Should prevent triggering if value is manually synced and cancelKey updated', async () => {
      const onTrigger = vi.fn();
      const lastValueRef = { v: '' };
      
      const { rerender } = render(Debounce, {
        value: 'test',
        onTrigger,
        delay: 500,
        lastValueRef,
        cancelKey: 0
      });
  
      vi.advanceTimersByTime(250);
      
      // Simulate manual search: update lastValueRef.v and cancelKey simultaneously
      lastValueRef.v = 'test';
      await rerender({
        value: 'test',
        onTrigger,
        delay: 500,
        lastValueRef,
        cancelKey: 1
      });
      
      vi.advanceTimersByTime(500);
      expect(onTrigger).not.toHaveBeenCalled();
  });

  it('Should cancel pending timer when component is destroyed', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: '' };
    
    render(Debounce, {
      value: 'test',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0
    });

    vi.advanceTimersByTime(250);
    cleanup(); // This destroys the component
    
    vi.advanceTimersByTime(500);
    expect(onTrigger).not.toHaveBeenCalled();
  });

  it('Should restart timer if value changes before delay', async () => {
    const onTrigger = vi.fn();
    const lastValueRef = { v: '' };
    
    const props = {
      value: 't',
      onTrigger,
      delay: 500,
      lastValueRef,
      cancelKey: 0
    };
    const { rerender } = render(Debounce, props);

    vi.advanceTimersByTime(200);
    props.value = 'te';
    await rerender(props);
    
    vi.advanceTimersByTime(200);
    props.value = 'tes';
    await rerender(props);

    vi.advanceTimersByTime(400);
    expect(onTrigger).not.toHaveBeenCalled();

    vi.advanceTimersByTime(100);
    expect(onTrigger).toHaveBeenCalledWith('tes');
  });
});
