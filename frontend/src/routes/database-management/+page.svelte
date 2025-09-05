<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isAuthenticated, connections } from '$lib/stores';
	import Input from '$lib/components/Input.svelte';

	// Main state
	let selectedConnection = null;
	let selectedCollection = null;
	let collections = [];
	let documents = [];
	let totalDocuments = 0;
	let loading = false;
	let error = '';
	let success = '';

	// Pagination
	let currentPage = 1;
	let pageSize = 10;
	let totalPages = 1;

	// Filters
	let searchQuery = '';
	let sortField = '';
	let sortOrder = 'asc';
	let filters = {};
	let availableFields = [];

	// Modal states
	let showCreateModal = false;
	let showEditModal = false;
	let showDeleteModal = false;
	let showFilterModal = false;
	let currentDocument = {};
	let documentToDelete = null;

	// Drag functionality for modals
	let isDraggingCreate = false;
	let isDraggingEdit = false;
	let isDraggingFilter = false;
	let createModalElement;
	let editModalElement;
	let filterModalElement;
	let createDragStartX = 0;
	let createDragStartY = 0;
	let editDragStartX = 0;
	let editDragStartY = 0;
	let filterDragStartX = 0;
	let filterDragStartY = 0;
	let createCurrentX = 0;
	let createCurrentY = 0;
	let editCurrentX = 0;
	let editCurrentY = 0;
	let filterCurrentX = 0;
	let filterCurrentY = 0;

	onMount(async () => {
		if (!$isAuthenticated) {
			goto('/login');
			return;
		}

		// Get connection ID from URL params
		const connectionId = $page.url.searchParams.get('connection');
		if (connectionId) {
			selectedConnection = $connections.find(c => c.id === connectionId);
			if (selectedConnection) {
				await loadCollections();
			}
		}
	});

	async function loadCollections() {
		if (!selectedConnection) return;

		loading = true;
		try {
			const response = await fetch(`http://localhost:8081/api/database-management/collections?database_id=${selectedConnection.id}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				collections = await response.json();
			} else {
				error = 'Failed to load collections';
			}
		} catch (err) {
			error = 'Error connecting to server';
		} finally {
			loading = false;
		}
	}

	async function selectCollection(collection) {
		selectedCollection = collection;
		currentPage = 1;
		await loadDocuments();
		await loadFieldSchema();
	}

	async function loadFieldSchema() {
		if (!selectedConnection || !selectedCollection) return;

		try {
			const response = await fetch(`http://localhost:8081/api/database-management/collections/${selectedCollection}/schema?database_id=${selectedConnection.id}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const schema = await response.json();
				availableFields = schema.fields || [];
			}
		} catch (err) {
			console.error('Error loading schema:', err);
		}
	}

	async function loadDocuments() {
		if (!selectedConnection || !selectedCollection) return;

		loading = true;
		try {
			const queryParams = new URLSearchParams({
				database_id: selectedConnection.id,
				page: currentPage.toString(),
				limit: pageSize.toString()
			});

			if (searchQuery) queryParams.append('search', searchQuery);
			if (sortField) {
				queryParams.append('sort', sortField);
				queryParams.append('order', sortOrder);
			}

			// Add filters
			Object.entries(filters).forEach(([key, value]) => {
				if (value) queryParams.append(`filter_${key}`, value);
			});

			const response = await fetch(`http://localhost:8081/api/database-management/collections/${selectedCollection}/documents?${queryParams}`, {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const result = await response.json();
				documents = result.documents || [];
				totalDocuments = result.total || 0;
				totalPages = Math.ceil(totalDocuments / pageSize);
			} else {
				error = 'Failed to load documents';
			}
		} catch (err) {
			error = 'Error connecting to server';
		} finally {
			loading = false;
		}
	}

	async function createDocument() {
		if (!selectedConnection || !selectedCollection) return;

		loading = true;
		try {
			const response = await fetch(`http://localhost:8081/api/database-management/collections/${selectedCollection}/documents`, {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					database_id: selectedConnection.id,
					data: currentDocument
				})
			});

			if (response.ok) {
				success = 'Document created successfully';
				showCreateModal = false;
				currentDocument = {};
				await loadDocuments();
			} else {
				const errorData = await response.json();
				error = errorData.error || 'Failed to create document';
			}
		} catch (err) {
			error = 'Error connecting to server';
		} finally {
			loading = false;
		}
	}

	async function updateDocument() {
		if (!selectedConnection || !selectedCollection || !currentDocument.id) return;

		loading = true;
		try {
			const response = await fetch(`http://localhost:8081/api/database-management/collections/${selectedCollection}/documents/${currentDocument.id}`, {
				method: 'PUT',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					database_id: selectedConnection.id,
					data: currentDocument
				})
			});

			if (response.ok) {
				success = 'Document updated successfully';
				showEditModal = false;
				await loadDocuments();
			} else {
				const errorData = await response.json();
				error = errorData.error || 'Failed to update document';
			}
		} catch (err) {
			error = 'Error connecting to server';
		} finally {
			loading = false;
		}
	}

	async function deleteDocument() {
		if (!selectedConnection || !selectedCollection || !documentToDelete) return;

		loading = true;
		try {
			const response = await fetch(`http://localhost:8081/api/database-management/collections/${selectedCollection}/documents/${documentToDelete.id}`, {
				method: 'DELETE',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					database_id: selectedConnection.id
				})
			});

			if (response.ok) {
				success = 'Document deleted successfully';
				showDeleteModal = false;
				documentToDelete = null;
				await loadDocuments();
			} else {
				const errorData = await response.json();
				error = errorData.error || 'Failed to delete document';
			}
		} catch (err) {
			error = 'Error connecting to server';
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		currentDocument = {};
		showCreateModal = true;
	}

	function openEditModal(document) {
		currentDocument = { ...document };
		showEditModal = true;
	}

	function openDeleteModal(document) {
		documentToDelete = document;
		showDeleteModal = true;
	}

	function closeCreateModal() {
		showCreateModal = false;
		currentDocument = {};
		resetCreateModalPosition();
	}

	function closeEditModal() {
		showEditModal = false;
		currentDocument = {};
		resetEditModalPosition();
	}

	function closeDeleteModal() {
		showDeleteModal = false;
		documentToDelete = null;
	}

	function closeFilterModal() {
		showFilterModal = false;
		resetFilterModalPosition();
	}

	function applyFilters() {
		currentPage = 1;
		loadDocuments();
		closeFilterModal();
	}

	function clearFilters() {
		filters = {};
		searchQuery = '';
		sortField = '';
		sortOrder = 'asc';
		currentPage = 1;
		loadDocuments();
		closeFilterModal();
	}

	function changePage(page) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
			loadDocuments();
		}
	}

	function handleSearch() {
		currentPage = 1;
		loadDocuments();
	}

	function handleLogout() {
		localStorage.removeItem('token');
		$isAuthenticated = false;
		goto('/login');
	}

	function formatValue(value) {
		if (value === null || value === undefined) return '-';
		if (typeof value === 'object') return JSON.stringify(value);
		return String(value);
	}

	function getDocumentKeys(docs) {
		if (!docs || docs.length === 0) return [];
		const keys = new Set();
		docs.forEach(doc => {
			Object.keys(doc).forEach(key => keys.add(key));
		});
		return Array.from(keys).filter(key => key !== '_id');
	}

	// Drag functionality for Create Modal
	function startCreateDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingCreate = true;
			
			const rect = createModalElement.getBoundingClientRect();
			createDragStartX = e.clientX - rect.left;
			createDragStartY = e.clientY - rect.top;
			
			createCurrentX = rect.left;
			createCurrentY = rect.top;
			
			createModalElement.style.left = createCurrentX + 'px';
			createModalElement.style.top = createCurrentY + 'px';
			createModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragCreate);
			document.addEventListener('mouseup', stopCreateDrag);
			
			e.preventDefault();
			createModalElement.style.userSelect = 'none';
		}
	}

	function dragCreate(e) {
		if (!isDraggingCreate) return;
		
		createCurrentX = e.clientX - createDragStartX;
		createCurrentY = e.clientY - createDragStartY;
		
		const modalRect = createModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		const margin = 50;
		createCurrentX = Math.max(-modalRect.width + margin, Math.min(createCurrentX, viewportWidth - margin));
		createCurrentY = Math.max(0, Math.min(createCurrentY, viewportHeight - modalRect.height));
		
		createModalElement.style.left = createCurrentX + 'px';
		createModalElement.style.top = createCurrentY + 'px';
	}

	function stopCreateDrag() {
		isDraggingCreate = false;
		createModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragCreate);
		document.removeEventListener('mouseup', stopCreateDrag);
	}

	function resetCreateModalPosition() {
		if (createModalElement) {
			createModalElement.style.left = '';
			createModalElement.style.top = '';
			createModalElement.style.transform = '';
			createCurrentX = 0;
			createCurrentY = 0;
		}
	}

	// Drag functionality for Edit Modal
	function startEditDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingEdit = true;
			
			const rect = editModalElement.getBoundingClientRect();
			editDragStartX = e.clientX - rect.left;
			editDragStartY = e.clientY - rect.top;
			
			editCurrentX = rect.left;
			editCurrentY = rect.top;
			
			editModalElement.style.left = editCurrentX + 'px';
			editModalElement.style.top = editCurrentY + 'px';
			editModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragEdit);
			document.addEventListener('mouseup', stopEditDrag);
			
			e.preventDefault();
			editModalElement.style.userSelect = 'none';
		}
	}

	function dragEdit(e) {
		if (!isDraggingEdit) return;
		
		editCurrentX = e.clientX - editDragStartX;
		editCurrentY = e.clientY - editDragStartY;
		
		const modalRect = editModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		const margin = 50;
		editCurrentX = Math.max(-modalRect.width + margin, Math.min(editCurrentX, viewportWidth - margin));
		editCurrentY = Math.max(0, Math.min(editCurrentY, viewportHeight - modalRect.height));
		
		editModalElement.style.left = editCurrentX + 'px';
		editModalElement.style.top = editCurrentY + 'px';
	}

	function stopEditDrag() {
		isDraggingEdit = false;
		editModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragEdit);
		document.removeEventListener('mouseup', stopEditDrag);
	}

	function resetEditModalPosition() {
		if (editModalElement) {
			editModalElement.style.left = '';
			editModalElement.style.top = '';
			editModalElement.style.transform = '';
			editCurrentX = 0;
			editCurrentY = 0;
		}
	}

	// Drag functionality for Filter Modal
	function startFilterDrag(e) {
		if (e.target.classList.contains('modal-header') || e.target.closest('.modal-header')) {
			isDraggingFilter = true;
			
			const rect = filterModalElement.getBoundingClientRect();
			filterDragStartX = e.clientX - rect.left;
			filterDragStartY = e.clientY - rect.top;
			
			filterCurrentX = rect.left;
			filterCurrentY = rect.top;
			
			filterModalElement.style.left = filterCurrentX + 'px';
			filterModalElement.style.top = filterCurrentY + 'px';
			filterModalElement.style.transform = 'none';
			
			document.addEventListener('mousemove', dragFilter);
			document.addEventListener('mouseup', stopFilterDrag);
			
			e.preventDefault();
			filterModalElement.style.userSelect = 'none';
		}
	}

	function dragFilter(e) {
		if (!isDraggingFilter) return;
		
		filterCurrentX = e.clientX - filterDragStartX;
		filterCurrentY = e.clientY - filterDragStartY;
		
		const modalRect = filterModalElement.getBoundingClientRect();
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		
		const margin = 50;
		filterCurrentX = Math.max(-modalRect.width + margin, Math.min(filterCurrentX, viewportWidth - margin));
		filterCurrentY = Math.max(0, Math.min(filterCurrentY, viewportHeight - modalRect.height));
		
		filterModalElement.style.left = filterCurrentX + 'px';
		filterModalElement.style.top = filterCurrentY + 'px';
	}

	function stopFilterDrag() {
		isDraggingFilter = false;
		filterModalElement.style.userSelect = '';
		document.removeEventListener('mousemove', dragFilter);
		document.removeEventListener('mouseup', stopFilterDrag);
	}

	function resetFilterModalPosition() {
		if (filterModalElement) {
			filterModalElement.style.left = '';
			filterModalElement.style.top = '';
			filterModalElement.style.transform = '';
			filterCurrentX = 0;
			filterCurrentY = 0;
		}
	}

	$: {
		if (error || success) {
			setTimeout(() => {
				error = '';
				success = '';
			}, 5000);
		}
	}
</script>

<svelte:head>
	<title>Database Management - Database Manager</title>
</svelte:head>

<nav class="navbar">
	<div class="container">
		<a href="/dashboard" class="navbar-brand">Database Manager</a>
		<div class="navbar-nav">
			<a href="/dashboard">Dashboard</a>
			<a href="/connections">Connections</a>
			<a href="/api-management">API Management</a>
			<a href="/database-management" class="active">Database Management</a>
			<button class="btn btn-secondary" on:click={handleLogout}>Logout</button>
		</div>
	</div>
</nav>

<div class="container">
	<div class="page-header">
		<h1>Database Management</h1>
		{#if selectedConnection}
			<p class="subtitle">Managing: <strong>{selectedConnection.name}</strong> ({selectedConnection.type})</p>
		{/if}
	</div>

	{#if error}
		<div class="alert alert-error">
			{error}
		</div>
	{/if}

	{#if success}
		<div class="alert alert-success">
			{success}
		</div>
	{/if}

	{#if !selectedConnection}
		<div class="no-connection">
			<h2>Select a Database Connection</h2>
			<p>Choose a database connection to manage collections and documents.</p>
			<div class="connections-grid">
				{#each $connections as connection}
					<div class="connection-card" on:click={() => {selectedConnection = connection; loadCollections()}}>
						<div class="connection-icon">
							{#if connection.type === 'mongodb'}
								🍃
							{:else if connection.type === 'mysql'}
								🐬
							{:else if connection.type === 'postgresql'}
								🐘
							{:else}
								💾
							{/if}
						</div>
						<h3>{connection.name}</h3>
						<p>{connection.type}</p>
						<p class="host">{connection.host}:{connection.port}</p>
					</div>
				{/each}
			</div>
		</div>
	{:else}
		<div class="database-management">
			<!-- Collections Sidebar -->
			<div class="sidebar">
				<div class="sidebar-header">
					<h3>Collections</h3>
					<button class="btn btn-sm" on:click={loadCollections} disabled={loading}>
						{loading ? '⟳' : '🔄'}
					</button>
				</div>
				<div class="collections-list">
					{#each collections as collection}
						<div 
							class="collection-item" 
							class:active={selectedCollection === collection}
							on:click={() => selectCollection(collection)}
						>
							📁 {collection}
						</div>
					{/each}
					{#if collections.length === 0 && !loading}
						<p class="no-collections">No collections found</p>
					{/if}
				</div>
			</div>

			<!-- Main Content -->
			<div class="main-content">
				{#if selectedCollection}
					<!-- Controls -->
					<div class="controls">
						<div class="controls-left">
							<h2>Collection: {selectedCollection}</h2>
							<p class="document-count">
								{totalDocuments} document{totalDocuments !== 1 ? 's' : ''}
							</p>
						</div>
						<div class="controls-right">
							<button class="btn btn-primary" on:click={openCreateModal}>
								➕ Add Document
							</button>
							<button class="btn btn-secondary" on:click={() => showFilterModal = true}>
								🔍 Filter
							</button>
							<button class="btn btn-secondary" on:click={loadDocuments} disabled={loading}>
								{loading ? '⟳' : '🔄'} Refresh
							</button>
						</div>
					</div>

					<!-- Search -->
					<div class="search-bar">
						<input
							type="text"
							placeholder="Search documents..."
							bind:value={searchQuery}
							on:input={handleSearch}
							class="search-input"
						/>
						{#if searchQuery}
							<button class="btn-clear" on:click={() => {searchQuery = ''; handleSearch();}}>×</button>
						{/if}
					</div>

					<!-- Documents Table -->
					{#if loading}
						<div class="loading">Loading documents...</div>
					{:else if documents.length > 0}
						<div class="table-container">
							<table class="documents-table">
								<thead>
									<tr>
										{#each getDocumentKeys(documents) as key}
											<th>
												<button 
													class="sort-header"
													on:click={() => {
														if (sortField === key) {
															sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
														} else {
															sortField = key;
															sortOrder = 'asc';
														}
														loadDocuments();
													}}
												>
													{key}
													{#if sortField === key}
														{sortOrder === 'asc' ? '↑' : '↓'}
													{/if}
												</button>
											</th>
										{/each}
										<th>Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each documents as document}
										<tr>
											{#each getDocumentKeys(documents) as key}
												<td>{formatValue(document[key])}</td>
											{/each}
											<td>
												<div class="action-buttons">
													<button 
														class="btn btn-sm btn-info"
														on:click={() => openEditModal(document)}
													>
														✏️ Edit
													</button>
													<button 
														class="btn btn-sm btn-danger"
														on:click={() => openDeleteModal(document)}
													>
														🗑️ Delete
													</button>
												</div>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>

						<!-- Pagination -->
						{#if totalPages > 1}
							<div class="pagination">
								<button 
									class="btn btn-sm"
									disabled={currentPage === 1}
									on:click={() => changePage(currentPage - 1)}
								>
									← Previous
								</button>
								
								<div class="page-numbers">
									{#each Array(Math.min(5, totalPages)) as _, i}
										{@const pageNum = Math.max(1, Math.min(totalPages - 4, currentPage - 2)) + i}
										{#if pageNum <= totalPages}
											<button 
												class="btn btn-sm"
												class:active={pageNum === currentPage}
												on:click={() => changePage(pageNum)}
											>
												{pageNum}
											</button>
										{/if}
									{/each}
								</div>
								
								<button 
									class="btn btn-sm"
									disabled={currentPage === totalPages}
									on:click={() => changePage(currentPage + 1)}
								>
									Next →
								</button>
								
								<div class="pagination-info">
									Page {currentPage} of {totalPages} 
									({(currentPage - 1) * pageSize + 1}-{Math.min(currentPage * pageSize, totalDocuments)} of {totalDocuments})
								</div>
							</div>
						{/if}
					{:else}
						<div class="no-documents">
							<h3>No documents found</h3>
							<p>This collection is empty or no documents match your search criteria.</p>
							<button class="btn btn-primary" on:click={openCreateModal}>
								➕ Add First Document
							</button>
						</div>
					{/if}
				{:else}
					<div class="no-collection">
						<h2>Select a Collection</h2>
						<p>Choose a collection from the sidebar to view and manage documents.</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Create Document Modal -->
{#if showCreateModal}
	<div 
		class="modal-content" 
		on:click|stopPropagation
		bind:this={createModalElement}
		class:dragging={isDraggingCreate}
	>
		<div 
			class="modal-header"
			on:mousedown={startCreateDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Add New Document
			</h2>
			<button class="modal-close" on:click={closeCreateModal}>&times;</button>
		</div>

		<form on:submit|preventDefault={createDocument}>
			<div class="form-group">
				<label class="form-label">Document Data (JSON)</label>
				<textarea
					bind:value={currentDocument}
					class="form-textarea"
					placeholder={'{"name": "John Doe", "email": "john@example.com"}'}
					rows="8"
					required
				></textarea>
				<small class="form-help">Enter valid JSON data for the document</small>
			</div>

			<div class="form-actions">
				<button type="button" class="btn btn-secondary" on:click={closeCreateModal}>
					Cancel
				</button>
				<button type="submit" class="btn btn-primary" disabled={loading}>
					{loading ? 'Creating...' : 'Create Document'}
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- Edit Document Modal -->
{#if showEditModal}
	<div 
		class="modal-content" 
		on:click|stopPropagation
		bind:this={editModalElement}
		class:dragging={isDraggingEdit}
	>
		<div 
			class="modal-header"
			on:mousedown={startEditDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Edit Document
			</h2>
			<button class="modal-close" on:click={closeEditModal}>&times;</button>
		</div>

		<form on:submit|preventDefault={updateDocument}>
			{#each Object.entries(currentDocument) as [key, value]}
				{#if key !== 'id' && key !== '_id'}
					<div class="form-group">
						<label class="form-label">{key}</label>
						<Input
							type="text"
							bind:value={currentDocument[key]}
							placeholder="Enter {key}"
							class="form-input"
						/>
					</div>
				{/if}
			{/each}

			<div class="form-actions">
				<button type="button" class="btn btn-secondary" on:click={closeEditModal}>
					Cancel
				</button>
				<button type="submit" class="btn btn-primary" disabled={loading}>
					{loading ? 'Updating...' : 'Update Document'}
				</button>
			</div>
		</form>
	</div>
{/if}

<!-- Delete Document Modal -->
{#if showDeleteModal}
	<div class="modal-content delete-modal">
		<div class="modal-header">
			<h2 class="modal-title">Confirm Delete</h2>
			<button class="modal-close" on:click={closeDeleteModal}>&times;</button>
		</div>

		<div class="modal-body">
			<p>Are you sure you want to delete this document?</p>
			<div class="document-preview">
				<pre>{JSON.stringify(documentToDelete, null, 2)}</pre>
			</div>
			<p class="warning">This action cannot be undone.</p>
		</div>

		<div class="form-actions">
			<button type="button" class="btn btn-secondary" on:click={closeDeleteModal}>
				Cancel
			</button>
			<button type="button" class="btn btn-danger" on:click={deleteDocument} disabled={loading}>
				{loading ? 'Deleting...' : 'Delete Document'}
			</button>
		</div>
	</div>
{/if}

<!-- Filter Modal -->
{#if showFilterModal}
	<div 
		class="modal-content filter-modal" 
		on:click|stopPropagation
		bind:this={filterModalElement}
		class:dragging={isDraggingFilter}
	>
		<div 
			class="modal-header"
			on:mousedown={startFilterDrag}
		>
			<h2 class="modal-title">
				<span class="drag-icon">⋮⋮</span>
				Filter Documents
			</h2>
			<button class="modal-close" on:click={closeFilterModal}>&times;</button>
		</div>

		<div class="modal-body">
			<div class="form-group">
				<label class="form-label">Sort Field</label>
				<select bind:value={sortField} class="form-select">
					<option value="">No sorting</option>
					{#each availableFields as field}
						<option value={field}>{field}</option>
					{/each}
				</select>
			</div>

			<div class="form-group">
				<label class="form-label">Sort Order</label>
				<select bind:value={sortOrder} class="form-select">
					<option value="asc">Ascending</option>
					<option value="desc">Descending</option>
				</select>
			</div>

			<div class="form-group">
				<label class="form-label">Filters</label>
				{#each availableFields as field}
					<div class="filter-field">
						<label>{field}</label>
						<Input
							type="text"
							bind:value={filters[field]}
							placeholder="Filter by {field}"
							class="form-input"
						/>
					</div>
				{/each}
			</div>
		</div>

		<div class="form-actions">
			<button type="button" class="btn btn-secondary" on:click={clearFilters}>
				Clear All
			</button>
			<button type="button" class="btn btn-secondary" on:click={closeFilterModal}>
				Cancel
			</button>
			<button type="button" class="btn btn-primary" on:click={applyFilters}>
				Apply Filters
			</button>
		</div>
	</div>
{/if}

<style>
	/* Main Layout */
	.container {
		max-width: 1400px;
		margin: 0 auto;
		padding: 20px;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: #333;
		margin: 0;
	}

	.subtitle {
		font-size: 1.1rem;
		color: #666;
		margin: 8px 0 0 0;
	}

	/* Navigation */
	.navbar {
		background: linear-gradient(135deg, #667eea, #764ba2);
		padding: 1rem 0;
		margin-bottom: 2rem;
	}

	.navbar .container {
		display: flex;
		justify-content: space-between;
		align-items: center;
		max-width: 1200px;
	}

	.navbar-brand {
		font-size: 1.5rem;
		font-weight: 700;
		color: white;
		text-decoration: none;
	}

	.navbar-nav {
		display: flex;
		gap: 2rem;
		align-items: center;
	}

	.navbar-nav a {
		color: rgba(255, 255, 255, 0.9);
		text-decoration: none;
		font-weight: 500;
		transition: color 0.2s;
	}

	.navbar-nav a:hover,
	.navbar-nav a.active {
		color: white;
		text-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
	}

	/* No Connection State */
	.no-connection {
		text-align: center;
		padding: 3rem;
	}

	.connections-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1.5rem;
		margin-top: 2rem;
	}

	.connection-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
		border: 2px solid #f1f5f9;
		cursor: pointer;
		transition: all 0.2s;
		text-align: center;
	}

	.connection-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
		border-color: #667eea;
	}

	.connection-icon {
		font-size: 2.5rem;
		margin-bottom: 1rem;
	}

	.connection-card h3 {
		margin: 0 0 0.5rem 0;
		color: #333;
	}

	.connection-card p {
		margin: 0.25rem 0;
		color: #666;
	}

	.host {
		font-family: monospace;
		font-size: 0.9rem;
		color: #888;
	}

	/* Database Management Layout */
	.database-management {
		display: grid;
		grid-template-columns: 300px 1fr;
		gap: 2rem;
		min-height: 600px;
	}

	/* Sidebar */
	.sidebar {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
		border: 2px solid #f1f5f9;
	}

	.sidebar-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #eee;
	}

	.sidebar-header h3 {
		margin: 0;
		color: #333;
	}

	.collections-list {
		max-height: 500px;
		overflow-y: auto;
	}

	.collection-item {
		padding: 0.75rem;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s;
		margin-bottom: 0.5rem;
		border: 1px solid transparent;
	}

	.collection-item:hover {
		background: #f8f9fa;
		border-color: #e9ecef;
	}

	.collection-item.active {
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border-color: #667eea;
	}

	.no-collections {
		text-align: center;
		color: #666;
		padding: 2rem;
	}

	/* Main Content */
	.main-content {
		background: white;
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
		border: 2px solid #f1f5f9;
	}

	.controls {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #eee;
	}

	.controls-left h2 {
		margin: 0 0 0.5rem 0;
		color: #333;
	}

	.document-count {
		margin: 0;
		color: #666;
		font-size: 0.9rem;
	}

	.controls-right {
		display: flex;
		gap: 0.5rem;
	}

	/* Search Bar */
	.search-bar {
		position: relative;
		margin-bottom: 1.5rem;
	}

	.search-input {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 2px solid #e2e8f0;
		border-radius: 8px;
		font-size: 1rem;
		transition: border-color 0.2s;
	}

	.search-input:focus {
		outline: none;
		border-color: #667eea;
	}

	.btn-clear {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		font-size: 1.2rem;
		cursor: pointer;
		color: #666;
	}

	/* Table */
	.table-container {
		overflow-x: auto;
		margin-bottom: 1.5rem;
	}

	.documents-table {
		width: 100%;
		border-collapse: collapse;
		background: white;
	}

	.documents-table th,
	.documents-table td {
		padding: 0.75rem;
		text-align: left;
		border-bottom: 1px solid #e2e8f0;
	}

	.documents-table th {
		background: #f8f9fa;
		font-weight: 600;
		color: #4a5568;
	}

	.sort-header {
		background: none;
		border: none;
		padding: 0;
		font-weight: 600;
		color: #4a5568;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.sort-header:hover {
		color: #667eea;
	}

	.documents-table td {
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.5rem;
		margin-top: 1.5rem;
	}

	.page-numbers {
		display: flex;
		gap: 0.25rem;
	}

	.pagination-info {
		margin-left: 1rem;
		color: #666;
		font-size: 0.9rem;
	}

	/* No Data States */
	.no-collection,
	.no-documents {
		text-align: center;
		padding: 3rem;
		color: #666;
	}

	.no-collection h2,
	.no-documents h3 {
		color: #333;
		margin-bottom: 1rem;
	}

	/* Action Buttons */
	.action-buttons {
		display: flex;
		gap: 0.5rem;
		justify-content: flex-start;
	}

	/* Buttons */
	.btn {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 6px;
		font-size: 0.9rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.8rem;
	}

	.btn-primary {
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
	}

	.btn-primary:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
	}

	.btn-secondary {
		background: #6c757d;
		color: white;
	}

	.btn-secondary:hover {
		background: #5a6268;
	}

	.btn-info {
		background: #17a2b8;
		color: white;
	}

	.btn-info:hover {
		background: #138496;
	}

	.btn-danger {
		background: #dc3545;
		color: white;
	}

	.btn-danger:hover {
		background: #c82333;
	}

	.btn.active {
		background: #667eea;
		color: white;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	/* Modal Styles */
	:global(.modal-content) {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: white;
		border-radius: 16px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
		width: 600px;
		max-width: 90vw;
		max-height: 90vh;
		overflow-y: auto;
		animation: modalSlideIn 0.3s ease-out;
		cursor: default;
		transition: box-shadow 0.2s ease;
		z-index: 1000;
		border: 2px solid rgba(102, 126, 234, 0.2);
	}

	:global(.modal-content.dragging) {
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
		transform: none !important;
		animation: none;
		z-index: 1001;
		border-color: rgba(102, 126, 234, 0.5);
	}

	:global(.modal-header) {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24px 24px 16px 24px;
		border-bottom: 1px solid #eee;
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border-radius: 16px 16px 0 0;
		cursor: move;
		user-select: none;
		position: relative;
	}

	:global(.modal-header:hover) {
		background: linear-gradient(135deg, #5a6fd8, #6a42a0);
	}

	:global(.modal-header:active) {
		cursor: grabbing;
	}

	:global(.modal-title) {
		font-size: 1.5rem;
		font-weight: 600;
		color: white;
		margin: 0;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	:global(.drag-icon) {
		font-size: 1.2rem;
		opacity: 0.7;
		cursor: move;
		transform: rotate(90deg);
		transition: opacity 0.2s;
	}

	:global(.modal-header:hover .drag-icon) {
		opacity: 1;
	}

	:global(.modal-close) {
		background: rgba(255, 255, 255, 0.2);
		border: none;
		color: white;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		font-size: 1.5rem;
		transition: background 0.2s;
	}

	:global(.modal-close:hover) {
		background: rgba(255, 255, 255, 0.3);
	}

	.modal-body {
		padding: 24px;
	}

	/* Form Styles */
	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 600;
		color: #374151;
	}

	.form-input,
	.form-select,
	.form-textarea {
		width: 100%;
		padding: 0.75rem;
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		font-size: 1rem;
		transition: border-color 0.2s, box-shadow 0.2s;
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #667eea;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	.form-textarea {
		resize: vertical;
		font-family: 'Fira Code', monospace;
	}

	.form-help {
		display: block;
		margin-top: 0.25rem;
		font-size: 0.875rem;
		color: #6b7280;
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		padding: 0 24px 24px;
	}

	.filter-field {
		margin-bottom: 1rem;
	}

	.filter-field label {
		display: block;
		margin-bottom: 0.25rem;
		font-size: 0.9rem;
		color: #666;
	}

	/* Document Preview */
	.document-preview {
		background: #f8f9fa;
		border-radius: 8px;
		padding: 1rem;
		margin: 1rem 0;
	}

	.document-preview pre {
		margin: 0;
		font-size: 0.9rem;
		max-height: 200px;
		overflow-y: auto;
	}

	.warning {
		color: #dc3545;
		font-weight: 500;
		text-align: center;
	}

	/* Delete Modal */
	.delete-modal {
		width: 500px;
	}

	.filter-modal {
		width: 500px;
	}

	/* Loading */
	.loading {
		text-align: center;
		padding: 2rem;
		color: #666;
	}

	/* Alerts */
	.alert {
		padding: 1rem;
		border-radius: 8px;
		margin-bottom: 1rem;
	}

	.alert-error {
		background: #fee2e2;
		color: #dc2626;
		border: 1px solid #fecaca;
	}

	.alert-success {
		background: #dcfce7;
		color: #16a34a;
		border: 1px solid #bbf7d0;
	}

	/* Animation */
	@keyframes modalSlideIn {
		from {
			opacity: 0;
			transform: translate(-50%, -50%) scale(0.95);
		}
		to {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
	}

	/* Responsive */
	@media (max-width: 768px) {
		.database-management {
			grid-template-columns: 1fr;
		}

		.sidebar {
			order: 2;
		}

		.main-content {
			order: 1;
		}

		.controls {
			flex-direction: column;
			gap: 1rem;
		}

		.controls-right {
			flex-wrap: wrap;
		}

		:global(.modal-content) {
			width: 95vw;
			margin: 20px;
		}
	}
</style>
