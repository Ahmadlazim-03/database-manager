<script>
  import { onMount } from 'svelte';
  
  export let open = false;
  export let title = '';
  export let size = 'md'; // sm, md, lg, xl
  
  let dialog;
  
  $: if (dialog && open) {
    dialog.showModal();
  }
  
  $: if (dialog && !open) {
    dialog.close();
  }
  
  function closeModal() {
    open = false;
  }
  
  function handleBackdropClick(event) {
    if (event.target === dialog) {
      closeModal();
    }
  }
  
  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeModal();
    }
  }
  
  onMount(() => {
    return () => {
      if (dialog && dialog.open) {
        dialog.close();
      }
    };
  });
</script>

<dialog
  bind:this={dialog}
  class="modal modal-{size}"
  on:click={handleBackdropClick}
  on:keydown={handleKeydown}
  data-aos="zoom-in"
  data-aos-duration="300"
>
  <div class="modal-content" on:click|stopPropagation>
    {#if title || $$slots.header}
      <div class="modal-header">
        {#if $$slots.header}
          <slot name="header" />
        {:else}
          <h3 class="modal-title">{title}</h3>
        {/if}
        <button class="modal-close" on:click={closeModal}>
          <i data-lucide="x"></i>
        </button>
      </div>
    {/if}
    
    <div class="modal-body">
      <slot />
    </div>
    
    {#if $$slots.footer}
      <div class="modal-footer">
        <slot name="footer" />
      </div>
    {/if}
  </div>
</dialog>

<style>
  dialog {
    padding: 0;
    border: none;
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-2xl);
    background: transparent;
    max-width: 90vw;
    max-height: 90vh;
  }
  
  dialog::backdrop {
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
  }
  
  .modal-content {
    background: var(--surface);
    border-radius: var(--radius-xl);
    overflow: hidden;
  }
  
  .modal-sm .modal-content {
    width: 400px;
  }
  
  .modal-md .modal-content {
    width: 500px;
  }
  
  .modal-lg .modal-content {
    width: 700px;
  }
  
  .modal-xl .modal-content {
    width: 900px;
  }
  
  .modal-header {
    padding: var(--space-6) var(--space-6) 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--divider-color);
    padding-bottom: var(--space-4);
    margin-bottom: var(--space-6);
  }
  
  .modal-title {
    font-size: var(--text-xl);
    font-weight: 700;
    color: var(--text-primary);
    margin: 0;
  }
  
  .modal-close {
    background: none;
    border: none;
    padding: var(--space-2);
    border-radius: var(--radius-md);
    cursor: pointer;
    color: var(--text-secondary);
    transition: var(--transition-colors);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .modal-close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  
  .modal-body {
    padding: var(--space-6);
    padding-top: 0;
  }
  
  .modal-footer {
    padding: 0 var(--space-6) var(--space-6);
    display: flex;
    gap: var(--space-3);
    justify-content: flex-end;
    border-top: 1px solid var(--divider-color);
    padding-top: var(--space-4);
    margin-top: var(--space-6);
  }
</style>
