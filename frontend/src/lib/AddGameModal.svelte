<script lang="ts">
  import { Modal, Button, Input, Label, Spinner, Checkbox } from 'flowbite-svelte';
  import { apiClient, type Game, type CreateGameRequest } from './api-client';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';

  export let open = false;
  export let gameId: string | null = null; // null for add, set for update
  export let onGameSaved: (game: Game) => void = () => {};
  export let onCancel: () => void = () => {};

  let gameTitle = '';
  let gameBarcode = '';
  let isPlayToWin = false;
  let barcodeLoading = false;
  let loading = false;

  // Determine if we're in edit mode
  $: isEditMode = gameId !== null;
  $: modalTitle = isEditMode ? 'Update Game' : 'Add Game';
  $: submitButtonText = isEditMode ? 'Update Game' : 'Add Game';

  // Reset and populate fields when modal opens
  $: if (open && gameId) {
    // Load existing game data for edit mode
    loadGameData();
  } else if (open && !gameId) {
    // Reset for add mode
    gameTitle = '';
    gameBarcode = '';
    isPlayToWin = false;
  } else if (!open) {
    // Clear when closed
    gameTitle = '';
    gameBarcode = '';
    isPlayToWin = false;
  }

  async function loadGameData() {
    if (!gameId) return;
    try {
      const game = await apiClient.getGame(gameId);
      gameTitle = game.title;
      gameBarcode = game.barcode || '';
      isPlayToWin = game.isPlayToWin;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load game';
      toasts.add(`Failed to load game: ${message}`, 'error');
    }
  }

  async function handleSubmit() {
    if (!gameTitle.trim()) return;

    loading = true;
    try {
      const requestBody: CreateGameRequest = {
        title: gameTitle.trim(),
        ...(gameBarcode.trim() ? { barcode: gameBarcode.trim() } : {}),
        isPlayToWin,
      };

      let savedGame: Game;

      if (isEditMode && gameId) {
        await apiClient.updateGame(gameId, requestBody);
        // After update, fetch the updated game to get the full object
        savedGame = await apiClient.getGame(gameId);
      } else {
        savedGame = await apiClient.addGame(requestBody);
      }

      onGameSaved(savedGame);
      open = false;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to save game';
      const action = isEditMode ? 'update' : 'add';
      toasts.add(`Failed to ${action} game: ${message}`, 'error');
    } finally {
      loading = false;
    }
  }

  function handleCancel() {
    onCancel();
    open = false;
  }

  // Suppress Enter on the title input — HID barcode scanners use Enter as their
  // terminator and would trigger submission before the librarian can review.
  function handleTitleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
    }
  }

  async function handleBarcodeKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      await handleBarcodeScan();
    }
  }

  async function handleBarcodeScan() {
    const value = gameBarcode.trim();
    if (!value) return;

    barcodeLoading = true;
    try {
      // Check if the barcode already belongs to an existing game
      const result = await apiClient.getGameByBarcode(value);
      if (result.games.length > 0) {
        // If we're in edit mode and this is the same game's barcode, allow it
        if (isEditMode && gameId && result.games[0].gameId === gameId) {
          // Same game, barcode is fine
          return;
        }
        // A different game already has this barcode
        toasts.add('A game with this barcode already exists', 'error');
        gameBarcode = '';
      }
    } catch (e) {
      // A 404 means the barcode is free to use — keep it in the field
      const message = e instanceof Error ? e.message : '';
      if (!message.toLowerCase().includes('not found') && !message.includes('404')) {
        toasts.add(`Barcode lookup failed: ${message}`, 'error');
        gameBarcode = '';
      }
      // else: barcode is available, leave it in the field
    } finally {
      barcodeLoading = false;
    }
  }
</script>

<Modal bind:open title={modalTitle} size="sm" autoclose={false}>
  <div class="space-y-4">
    <div>
      <Label for="addGameTitle" class="mb-2">Game Title</Label>
      <Input
        id="addGameTitle"
        placeholder="Enter game title"
        bind:value={gameTitle}
        onkeydown={handleTitleKeydown}
        autocomplete="off"
        maxlength={100}
        disabled={loading}
      />
    </div>

    {#if isBarcodeEnabled()}
      <div>
        <Label for="addGameBarcode" class="mb-2">
          <span
            class="text-xs font-medium tracking-wide text-slate-400 uppercase dark:text-slate-500"
          >
            Game Barcode
          </span>
        </Label>
        <div class="relative">
          <Input
            id="addGameBarcode"
            placeholder="Scan game barcode…"
            bind:value={gameBarcode}
            onkeydown={handleBarcodeKeydown}
            autocomplete="off"
            maxlength={48}
            disabled={loading || barcodeLoading}
          />
          {#if barcodeLoading}
            <div class="pointer-events-none absolute inset-y-0 inset-e-0 flex items-center pe-3">
              <Spinner size="4" />
            </div>
          {/if}
        </div>
        <p class="mt-1 text-xs text-slate-400 dark:text-slate-500">
          Optional. Scan a UPC or barcode to associate with this game.
        </p>
      </div>
    {/if}

    <div>
      <Checkbox
        bind:checked={isPlayToWin}
        disabled={loading}
        class="rounded border-slate-300 text-sky-600 focus:ring-sky-500 dark:border-slate-600 dark:bg-slate-700 dark:ring-offset-slate-800 dark:focus:ring-sky-600"
      >
        <Label class="ms-2 text-sm font-medium text-slate-900 dark:text-slate-100">
          Play to Win Game
        </Label>
      </Checkbox>
      <p class="mt-1 text-xs text-slate-400 dark:text-slate-500">
        Mark this game as a Play to Win game.
      </p>
    </div>

    <div class="flex justify-end gap-2 pt-2">
      <Button color="alternative" onclick={handleCancel} disabled={loading}>Cancel</Button>
      <Button
        data-testid="add-game-submit"
        onclick={handleSubmit}
        disabled={loading || barcodeLoading || !gameTitle.trim()}
      >
        {#if loading}
          <Spinner size="4" class="me-2" />
        {/if}
        {submitButtonText}
      </Button>
    </div>
  </div>
</Modal>

