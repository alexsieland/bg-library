import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createPageScrollHandler } from './pageScrollLoader';

function setupScrollEnv({
  scrollY = 0,
  innerHeight = 800,
  scrollHeight = 1000,
}: {
  scrollY?: number;
  innerHeight?: number;
  scrollHeight?: number;
} = {}) {
  Object.defineProperty(window, 'scrollY', { configurable: true, value: scrollY });
  Object.defineProperty(window, 'innerHeight', { configurable: true, value: innerHeight });
  Object.defineProperty(document.documentElement, 'scrollHeight', {
    configurable: true,
    value: scrollHeight,
  });
}

describe('createPageScrollHandler', () => {
  let loadMore: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    loadMore = vi.fn();
  });

  it('Should call loadMore when scrolled within threshold of the bottom', () => {
    // scrollHeight(1000) - (scrollY(150) + innerHeight(800)) = 50 < 100 threshold
    setupScrollEnv({ scrollY: 150, innerHeight: 800, scrollHeight: 1000 });

    const handler = createPageScrollHandler({
      isLoading: () => false,
      isLoadingMore: () => false,
      hasMore: () => true,
      hasError: () => false,
      loadMore,
    });

    handler();

    expect(loadMore).toHaveBeenCalledOnce();
  });

  it('Should not call loadMore when not near the bottom', () => {
    // scrollHeight(2000) - (scrollY(0) + innerHeight(800)) = 1200 > 100 threshold
    setupScrollEnv({ scrollY: 0, innerHeight: 800, scrollHeight: 2000 });

    const handler = createPageScrollHandler({
      isLoading: () => false,
      isLoadingMore: () => false,
      hasMore: () => true,
      hasError: () => false,
      loadMore,
    });

    handler();

    expect(loadMore).not.toHaveBeenCalled();
  });

  it('Should not call loadMore when isLoading is true', () => {
    setupScrollEnv({ scrollY: 150, innerHeight: 800, scrollHeight: 1000 });

    const handler = createPageScrollHandler({
      isLoading: () => true,
      isLoadingMore: () => false,
      hasMore: () => true,
      hasError: () => false,
      loadMore,
    });

    handler();

    expect(loadMore).not.toHaveBeenCalled();
  });

  it('Should not call loadMore when isLoadingMore is true', () => {
    setupScrollEnv({ scrollY: 150, innerHeight: 800, scrollHeight: 1000 });

    const handler = createPageScrollHandler({
      isLoading: () => false,
      isLoadingMore: () => true,
      hasMore: () => true,
      hasError: () => false,
      loadMore,
    });

    handler();

    expect(loadMore).not.toHaveBeenCalled();
  });

  it('Should not call loadMore when hasMore is false', () => {
    setupScrollEnv({ scrollY: 150, innerHeight: 800, scrollHeight: 1000 });

    const handler = createPageScrollHandler({
      isLoading: () => false,
      isLoadingMore: () => false,
      hasMore: () => false,
      hasError: () => false,
      loadMore,
    });

    handler();

    expect(loadMore).not.toHaveBeenCalled();
  });

  it('Should not call loadMore when hasError is true', () => {
    setupScrollEnv({ scrollY: 150, innerHeight: 800, scrollHeight: 1000 });

    const handler = createPageScrollHandler({
      isLoading: () => false,
      isLoadingMore: () => false,
      hasMore: () => true,
      hasError: () => true,
      loadMore,
    });

    handler();

    expect(loadMore).not.toHaveBeenCalled();
  });

  it('Should respect a custom threshold', () => {
    // scrollHeight(1000) - (scrollY(100) + innerHeight(800)) = 100, exactly at default threshold
    // With a threshold of 50 this should NOT fire
    setupScrollEnv({ scrollY: 100, innerHeight: 800, scrollHeight: 1000 });

    const handler = createPageScrollHandler({
      isLoading: () => false,
      isLoadingMore: () => false,
      hasMore: () => true,
      hasError: () => false,
      loadMore,
      threshold: 50,
    });

    handler();

    expect(loadMore).not.toHaveBeenCalled();
  });
});
