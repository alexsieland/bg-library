/**
 * Creates a page-level scroll handler for infinite-scroll tables.
 *
 * Bind the returned function to `<svelte:window on:scroll={handler} />`.
 * The handler fires `loadMore` when the user has scrolled within
 * `threshold` pixels of the page bottom, provided none of the guard
 * conditions are active.
 *
 * @param options.isLoading     - True while the initial/search load is in flight.
 * @param options.isLoadingMore - True while an incremental page is in flight.
 * @param options.hasMore       - False once the last page returned fewer items than PAGE_LIMIT.
 * @param options.hasError      - True when the last fetch produced an error.
 * @param options.loadMore      - Callback to trigger loading the next page.
 * @param options.threshold     - Pixels from the bottom that trigger loading (default 100).
 */
export function createPageScrollHandler(options: {
  isLoading: () => boolean;
  isLoadingMore: () => boolean;
  hasMore: () => boolean;
  hasError: () => boolean;
  loadMore: () => void;
  threshold?: number;
}): () => void {
  const threshold = options.threshold ?? 100;

  return function handlePageScroll() {
    if (options.isLoading() || options.isLoadingMore() || !options.hasMore() || options.hasError()) {
      return;
    }

    const scrollTop = window.scrollY || document.documentElement.scrollTop;
    const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
    const pageHeight = document.documentElement.scrollHeight;
    const isAtBottom = pageHeight - (scrollTop + viewportHeight) < threshold;

    if (isAtBottom) {
      options.loadMore();
    }
  };
}
