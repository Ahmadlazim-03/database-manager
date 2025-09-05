<script>
  export let label = '';
  export let value = '';
  export let type = 'text';
  export let placeholder = '';
  export let required = false;
  export let disabled = false;
  export let error = '';
  export let help = '';
  export let icon = null;
  export let floating = false;
  
  let inputElement;
  let focused = false;
  
  $: hasValue = value && value.length > 0;
  $: inputClasses = [
    'form-input',
    error && 'invalid',
    !error && hasValue && 'valid'
  ].filter(Boolean).join(' ');
  
  function handleFocus() {
    focused = true;
  }
  
  function handleBlur() {
    focused = false;
  }
</script>

<div class="form-group">
  {#if floating}
    <div class="floating-label">
      {#if icon}
        <div class="input-group">
          <i class="input-icon" data-lucide={icon}></i>
          {#if type === 'email'}
            <input
              bind:this={inputElement}
              bind:value
              type="email"
              {placeholder}
              {required}
              {disabled}
              class={inputClasses}
              on:focus={handleFocus}
              on:blur={handleBlur}
              on:input
              on:change
              style="padding-left: var(--space-12);"
            />
          {:else if type === 'password'}
            <input
              bind:this={inputElement}
              bind:value
              type="password"
              {placeholder}
              {required}
              {disabled}
              class={inputClasses}
              on:focus={handleFocus}
              on:blur={handleBlur}
              on:input
              on:change
              style="padding-left: var(--space-12);"
            />
          {:else if type === 'number'}
            <input
              bind:this={inputElement}
              bind:value
              type="number"
              {placeholder}
              {required}
              {disabled}
              class={inputClasses}
              on:focus={handleFocus}
              on:blur={handleBlur}
              on:input
              on:change
              style="padding-left: var(--space-12);"
            />
          {:else}
            <input
              bind:this={inputElement}
              bind:value
              type="text"
              {placeholder}
              {required}
              {disabled}
              class={inputClasses}
              on:focus={handleFocus}
              on:blur={handleBlur}
              on:input
              on:change
              style="padding-left: var(--space-12);"
            />
          {/if}
          <label class="form-label" class:required>{label}</label>
        </div>
      {:else}
        {#if type === 'email'}
          <input
            bind:this={inputElement}
            bind:value
            type="email"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {:else if type === 'password'}
          <input
            bind:this={inputElement}
            bind:value
            type="password"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {:else if type === 'number'}
          <input
            bind:this={inputElement}
            bind:value
            type="number"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {:else}
          <input
            bind:this={inputElement}
            bind:value
            type="text"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {/if}
        <label class="form-label" class:required>{label}</label>
      {/if}
    </div>
  {:else}
    {#if label}
      <label class="form-label" class:required for={inputElement}>{label}</label>
    {/if}
    
    {#if icon}
      <div class="input-group">
        <i class="input-icon" data-lucide={icon}></i>
        {#if type === 'email'}
          <input
            bind:this={inputElement}
            bind:value
            type="email"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {:else if type === 'password'}
          <input
            bind:this={inputElement}
            bind:value
            type="password"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {:else if type === 'number'}
          <input
            bind:this={inputElement}
            bind:value
            type="number"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {:else}
          <input
            bind:this={inputElement}
            bind:value
            type="text"
            {placeholder}
            {required}
            {disabled}
            class={inputClasses}
            on:focus={handleFocus}
            on:blur={handleBlur}
            on:input
            on:change
          />
        {/if}
      </div>
    {:else}
      {#if type === 'email'}
        <input
          bind:this={inputElement}
          bind:value
          type="email"
          {placeholder}
          {required}
          {disabled}
          class={inputClasses}
          on:focus={handleFocus}
          on:blur={handleBlur}
          on:input
          on:change
        />
      {:else if type === 'password'}
        <input
          bind:this={inputElement}
          bind:value
          type="password"
          {placeholder}
          {required}
          {disabled}
          class={inputClasses}
          on:focus={handleFocus}
          on:blur={handleBlur}
          on:input
          on:change
        />
      {:else if type === 'number'}
        <input
          bind:this={inputElement}
          bind:value
          type="number"
          {placeholder}
          {required}
          {disabled}
          class={inputClasses}
          on:focus={handleFocus}
          on:blur={handleBlur}
          on:input
          on:change
        />
      {:else}
        <input
          bind:this={inputElement}
          bind:value
          type="text"
          {placeholder}
          {required}
          {disabled}
          class={inputClasses}
          on:focus={handleFocus}
          on:blur={handleBlur}
          on:input
          on:change
        />
      {/if}
    {/if}
  {/if}
  
  {#if error}
    <div class="form-error">
      <i data-lucide="alert-circle"></i>
      {error}
    </div>
  {/if}
  
  {#if help && !error}
    <div class="form-help">{help}</div>
  {/if}
</div>

<style>
  /* Input styles are already defined in app.html */
</style>
