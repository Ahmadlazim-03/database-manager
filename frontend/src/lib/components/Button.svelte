<script>
  export let variant = 'primary'; // primary, secondary, success, warning, danger, ghost
  export let size = 'md'; // sm, md, lg
  export let loading = false;
  export let disabled = false;
  export let href = null;
  export let type = 'button';
  export let icon = null;
  export let iconPosition = 'left'; // left, right
  
  $: classes = [
    'btn',
    `btn-${variant}`,
    `btn-${size}`,
    loading && 'loading',
    disabled && 'disabled'
  ].filter(Boolean).join(' ');
  
  function handleClick(event) {
    if (loading || disabled) {
      event.preventDefault();
      return;
    }
    // Dispatch click event
  }
</script>

{#if href}
  <a {href} class={classes} on:click={handleClick} role="button" tabindex="0">
    {#if icon && iconPosition === 'left'}
      <i data-lucide={icon}></i>
    {/if}
    <slot />
    {#if icon && iconPosition === 'right'}
      <i data-lucide={icon}></i>
    {/if}
  </a>
{:else}
  <button {type} class={classes} {disabled} on:click={handleClick}>
    {#if icon && iconPosition === 'left'}
      <i data-lucide={icon}></i>
    {/if}
    <slot />
    {#if icon && iconPosition === 'right'}
      <i data-lucide={icon}></i>
    {/if}
  </button>
{/if}

<style>
  /* Button styles are already defined in app.html */
</style>
