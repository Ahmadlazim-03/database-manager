<script>
  export let type = 'success'; // success, error, warning, info
  export let title = '';
  export let message = '';
  export let dismissible = true;
  export let icon = null;
  
  let visible = true;
  
  const iconMap = {
    success: 'check-circle',
    error: 'x-circle', 
    warning: 'alert-triangle',
    info: 'info'
  };
  
  $: alertIcon = icon || iconMap[type];
  
  function dismiss() {
    visible = false;
  }
</script>

{#if visible}
  <div class="alert alert-{type}" data-aos="fade-in" data-aos-duration="300">
    <div class="alert-icon">
      <i data-lucide={alertIcon}></i>
    </div>
    <div class="alert-content">
      {#if title}
        <div class="alert-title">{title}</div>
      {/if}
      <div class="alert-message">
        {#if message}
          {message}
        {:else}
          <slot />
        {/if}
      </div>
    </div>
    {#if dismissible}
      <button class="btn-ghost btn-sm" on:click={dismiss}>
        <i data-lucide="x"></i>
      </button>
    {/if}
  </div>
{/if}

<style>
  /* Alert styles are already defined in app.html */
</style>
