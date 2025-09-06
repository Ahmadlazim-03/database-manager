<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isAuthenticated, connections } from '$lib/stores';
	import { apiClient } from '$lib/api';
	import { config } from '$lib/config.js';
	import Input from '$lib/components/Input.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import ShareDatabaseModal from '$lib/components/ShareDatabaseModal.svelte';

	// Main state
	let selectedConnection = null;
	let selectedCollection = null;
	let collections = [];
	let documents = [];
	let totalDocuments = 0;
	let loading = false;
	let error = '';
	let success = '';
	let sharedDatabases = [];

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
	let collectionSchema = []; // Schema for the current collection

	// Modal states
	let showCreateModal = false;
	let showEditModal = false;
	let showDeleteModal = false;
	let showFilterModal = false;
	let showShareModal = false;
	let currentDocument = {};
	let newDocumentData = {}; // For form-based creation
	let currentDocumentJSON = '';
	let documentToDelete = null;

	// Permission and sharing states
	let isSharedDatabase = false;
	let databasePermission = 'read'; // read, write, admin

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

		// Load both connections and shared databases first
		await Promise.all([
			loadConnections(),
			loadSharedDatabases()
		]);

		// Check if this is a shared database from URL params or localStorage
		const urlShared = $page.url.searchParams.get('shared');
		const urlPermission = $page.url.searchParams.get('permission');
		const storedShared = localStorage.getItem('isSharedDatabase');
		const storedPermission = localStorage.getItem('selectedDatabasePermission');

		if (urlShared === 'true' || storedShared === 'true') {
			isSharedDatabase = true;
			databasePermission = urlPermission || storedPermission || 'read';
		}

		// Get database ID from URL params (either 'connection' or 'db')
		const connectionId = $page.url.searchParams.get('connection');
		const dbId = $page.url.searchParams.get('db');
		const storedDbId = localStorage.getItem('selectedDatabaseId');
		const targetId = connectionId || dbId || storedDbId;
		
		if (targetId) {
			// First look in owned connections
			selectedConnection = $connections.find(c => c.id === targetId);
			
			// If not found, look in shared databases
			if (!selectedConnection) {
				const sharedDb = sharedDatabases.find(s => s.database.id === targetId);
				if (sharedDb) {
					selectedConnection = sharedDb.database;
					isSharedDatabase = true;
					databasePermission = sharedDb.permission_level;
				}
			}
			
			if (selectedConnection) {
				await loadCollections();
			}
		}

		// Clear localStorage after use
		localStorage.removeItem('selectedDatabaseId');
		localStorage.removeItem('selectedDatabasePermission');
		localStorage.removeItem('isSharedDatabase');
	});

	async function loadSharedDatabases() {
		try {
			console.log('📡 Fetching shared databases...');
			sharedDatabases = await apiClient.getSharedDatabases();
			console.log('✅ Received shared databases:', sharedDatabases);
		} catch (err) {
			console.error('❌ Failed to load shared databases:', err);
		}
	}

	async function loadConnections() {
		try {
			console.log('📡 Fetching connections from /api/database...');
			const response = await fetch(config.getApiUrl('/database'), {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const connectionsData = await response.json();
				console.log('✅ Received connections data:', connectionsData);
				connections.set(connectionsData);
			} else {
				console.error('❌ Failed to fetch connections:', response.status, response.statusText);
			}
		} catch (err) {
			console.error('❌ Error loading connections:', err);
		}
	}

	// Check if current user owns the selected database
	function isOwner(connection) {
		if (!connection) return false;
		// If connection is from $connections, user is owner
		return $connections.some(c => c.id === connection.id);
	}

	// Check if user can perform write operations
	function canWrite() {
		if (isOwner(selectedConnection)) return true;
		return isSharedDatabase && (databasePermission === 'write' || databasePermission === 'admin');
	}

	// Check if user can perform admin operations (like sharing)
	function canAdmin() {
		const isOwnerCheck = isOwner(selectedConnection);
		const hasAdminAccess = isSharedDatabase && databasePermission === 'admin';
		console.log('canAdmin check:', { isOwnerCheck, isSharedDatabase, databasePermission, hasAdminAccess });
		if (isOwnerCheck) return true;
		return hasAdminAccess;
	}

	async function leaveSharedDatabase() {
		if (!isSharedDatabase || !selectedConnection) return;

		if (!confirm('Are you sure you want to leave this shared database? You will lose access to it and will need to be re-invited.')) {
			return;
		}

		try {
			await apiClient.leaveSharedDatabase(selectedConnection.id);
			alert('Successfully left the shared database');
			
			// Redirect to dashboard
			goto('/dashboard');
		} catch (error) {
			console.error('Error leaving shared database:', error);
			alert('Failed to leave shared database: ' + (error.response?.data?.error || error.message));
		}
	}

	async function loadCollections() {
		if (!selectedConnection) return;

		loading = true;
		try {
			const response = await fetch(config.getApiUrl(`/database-management/collections?database_id=${selectedConnection.id}`), {
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
			const response = await fetch(config.getApiUrl(`/database-management/collections/${selectedCollection}/schema?database_id=${selectedConnection.id}`), {
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`
				}
			});

			if (response.ok) {
				const schema = await response.json();
				// Handle both old format (array of strings) and new format (array of objects)
				if (schema.fields && schema.fields.length > 0) {
					if (typeof schema.fields[0] === 'string') {
						// Old format: array of strings
						availableFields = schema.fields;
						collectionSchema = schema.fields.map(field => ({ name: field, type: 'text' }));
					} else {
						// New format: array of objects with name and type
						availableFields = schema.fields.map(field => field.name);
						collectionSchema = schema.fields;
					}
				} else {
					availableFields = [];
					collectionSchema = [];
				}
				
				console.log('Schema loaded for', selectedCollection, ':', collectionSchema);
				
				// Initialize newDocumentData with empty values
				initializeNewDocumentData();
				
				// Update template if create modal is open
				if (showCreateModal && (!currentDocumentJSON || currentDocumentJSON.trim() === '')) {
					currentDocumentJSON = generateJSONTemplate(availableFields);
					console.log('Generated template:', currentDocumentJSON);
				}
			}
		} catch (err) {
			console.error('Error loading schema:', err);
		}
	}

	function initializeNewDocumentData() {
		newDocumentData = {};
		collectionSchema.forEach(field => {
			// Handle both old format (string) and new format (object)
			const fieldName = typeof field === 'string' ? field : field.name;
			// Skip auto-generated fields like id, created_at, updated_at
			if (!['id', 'created_at', 'updated_at'].includes(fieldName.toLowerCase())) {
				newDocumentData[fieldName] = '';
			}
		});
	}

	// Handle image upload with compression
	async function handleImageUpload(event, fieldName) {
		const file = event.target.files[0];
		if (!file) return;

		// Validate file type
		const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
		if (!allowedTypes.includes(file.type)) {
			error = 'Please select a valid image file (JPEG, PNG, GIF, or WebP)';
			event.target.value = ''; // Clear the input
			return;
		}

		// Validate file size (max 10MB for original file)
		const maxOriginalSize = 10 * 1024 * 1024; // 10MB
		if (file.size > maxOriginalSize) {
			error = 'Original image file size must be less than 10MB';
			event.target.value = ''; // Clear the input
			return;
		}

		try {
			loading = true;
			error = null;

			// Convert file to base64
			const base64 = await convertFileToBase64(file);
			
			// Compress the image to reduce size for database storage
			const compressedBase64 = await compressImage(base64, 0.7, 800, 600);
			
			// Check final size (should be under 1MB for database storage)
			const finalSizeKB = Math.round((compressedBase64.length * 0.75) / 1024); // Approximate size in KB
			console.log(`Image processed for ${fieldName}: Original: ${Math.round(file.size/1024)}KB, Compressed: ${finalSizeKB}KB`);
			
			if (finalSizeKB > 1024) { // If still larger than 1MB, compress more
				const moreCompressed = await compressImage(base64, 0.5, 600, 400);
				newDocumentData[fieldName] = moreCompressed;
				console.log(`Further compressed to: ${Math.round((moreCompressed.length * 0.75) / 1024)}KB`);
			} else {
				newDocumentData[fieldName] = compressedBase64;
			}
			
		} catch (err) {
			console.error('Error uploading image:', err);
			error = 'Failed to process image: ' + err.message;
			event.target.value = ''; // Clear the input
		} finally {
			loading = false;
		}
	}

	// Convert file to base64
	function convertFileToBase64(file) {
		return new Promise((resolve, reject) => {
			const reader = new FileReader();
			reader.onload = () => resolve(reader.result);
			reader.onerror = () => reject(new Error('Failed to read file'));
			reader.readAsDataURL(file);
		});
	}

	// Compress image to reduce size
	function compressImage(base64, quality = 0.7, maxWidth = 800, maxHeight = 600) {
		return new Promise((resolve, reject) => {
			const canvas = document.createElement('canvas');
			const ctx = canvas.getContext('2d');
			const img = new Image();
			
			img.onload = () => {
				try {
					// Calculate new dimensions while maintaining aspect ratio
					let { width, height } = img;
					
					if (width > maxWidth || height > maxHeight) {
						const ratio = Math.min(maxWidth / width, maxHeight / height);
						width = Math.floor(width * ratio);
						height = Math.floor(height * ratio);
					}
					
					canvas.width = width;
					canvas.height = height;
					
					// Use better image rendering
					ctx.imageSmoothingEnabled = true;
					ctx.imageSmoothingQuality = 'high';
					
					// Draw and compress
					ctx.drawImage(img, 0, 0, width, height);
					const compressedBase64 = canvas.toDataURL('image/jpeg', quality);
					resolve(compressedBase64);
				} catch (error) {
					reject(error);
				}
			};
			
			img.onerror = () => reject(new Error('Failed to load image'));
			img.src = base64;
		});
	}

	// Remove uploaded image
	function removeImage(fieldName) {
		newDocumentData[fieldName] = '';
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

			const response = await fetch(config.getApiUrl(`/database-management/collections/${selectedCollection}/documents?${queryParams}`), {
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
			let documentData = {};
			
			// Use form data instead of JSON
			documentData = { ...newDocumentData };
			
			// Remove empty fields
			Object.keys(documentData).forEach(key => {
				if (documentData[key] === '' || documentData[key] === null || documentData[key] === undefined) {
					delete documentData[key];
				}
			});
			
			if (Object.keys(documentData).length === 0) {
				error = 'Please fill at least one field';
				loading = false;
				return;
			}
			
			console.log('Document data from form:', documentData);

			const requestBody = {
				database_id: selectedConnection.id,
				data: documentData
			};
			
			console.log('Sending request to create document:', requestBody);

			const response = await fetch(config.getApiUrl(`/database-management/collections/${selectedCollection}/documents`), {
				method: 'POST',
				headers: {
					'Authorization': `Bearer ${localStorage.getItem('token')}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(requestBody)
			});

			console.log('Response status:', response.status);
			
			if (response.ok) {
				const result = await response.json();
				console.log('Create success:', result);
				success = 'Document created successfully';
				showCreateModal = false;
				currentDocument = {};
				currentDocumentJSON = '';
				await loadDocuments();
			} else {
				const errorData = await response.json();
				console.error('Create error:', errorData);
				error = errorData.error || 'Failed to create document';
			}
		} catch (err) {
			console.error('Network error:', err);
			error = 'Error connecting to server';
		} finally {
			loading = false;
		}
	}

	async function updateDocument() {
		if (!selectedConnection || !selectedCollection || !currentDocument.id) return;

		loading = true;
		try {
			const response = await fetch(config.getApiUrl(`/database-management/collections/${selectedCollection}/documents/${currentDocument.id}`), {
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
			const response = await fetch(config.getApiUrl(`/database-management/collections/${selectedCollection}/documents/${documentToDelete.id}`), {
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

	async function openCreateModal() {
		// Load schema first
		await loadFieldSchema();
		// Initialize form data
		initializeNewDocumentData();
		showCreateModal = true;
		// Generate JSON template as fallback
		currentDocumentJSON = generateJSONTemplate(availableFields);
		console.log('openCreateModal - availableFields:', availableFields);
		console.log('openCreateModal - newDocumentData:', newDocumentData);
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
		newDocumentData = {};
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

	function generateJSONTemplate(fields) {
		if (!fields || fields.length === 0) {
			return '{\n  "name": "John Doe",\n  "email": "john@example.com"\n}';
		}

		const template = {};
		fields.forEach(field => {
			const fieldLower = field.toLowerCase();
			
			if (fieldLower.includes('email')) {
				const timestamp = Date.now();
				template[field] = `user${timestamp}@example.com`;
			} else if (fieldLower.includes('nama') || fieldLower.includes('name')) {
				template[field] = 'John Doe';
			} else if (fieldLower.includes('nim') || fieldLower.includes('nip')) {
				template[field] = '1234567890';
			} else if (fieldLower.includes('jurusan') || fieldLower.includes('department')) {
				template[field] = 'Computer Science';
			} else if (fieldLower.includes('angkatan') || fieldLower.includes('year')) {
				template[field] = 2024;
			} else if (fieldLower.includes('password')) {
				template[field] = 'password123';
			} else if (fieldLower.includes('age')) {
				template[field] = 25;
			} else if (fieldLower === 'id' || fieldLower.endsWith('_id')) {
				template[field] = Math.floor(Math.random() * 10000) + 1;
			} else if (fieldLower.includes('active') || fieldLower.includes('verified') || fieldLower.includes('is_')) {
				template[field] = true;
			} else if (fieldLower.includes('date') || fieldLower.includes('time') || fieldLower.includes('created') || fieldLower.includes('updated')) {
				template[field] = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
			} else if (fieldLower.includes('phone') || fieldLower.includes('mobile')) {
				template[field] = '+1234567890';
			} else if (fieldLower.includes('address')) {
				template[field] = '123 Main Street, City, Country';
			} else if (fieldLower.includes('role')) {
				template[field] = 'user';
			} else if (fieldLower.includes('status')) {
				template[field] = 'active';
			} else if (fieldLower.includes('count') || fieldLower.includes('number') || fieldLower.includes('price') || fieldLower.includes('amount')) {
				template[field] = 100;
			} else {
				// For unknown fields, use appropriate default based on common patterns
				if (fieldLower.length <= 5 && (fieldLower.includes('id') || /^\d+$/.test(fieldLower))) {
					template[field] = Math.floor(Math.random() * 1000) + 1;
				} else {
					template[field] = 'sample_value';
				}
			}
		});

		return JSON.stringify(template, null, 2);
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

<Navbar />

<div class="container">
	<div class="page-header">
		<h1>Database Management</h1>
		{#if selectedConnection}
			<div class="subtitle">
				<div class="subtitle-info">
					Managing: <strong>{selectedConnection.name}</strong> ({selectedConnection.type})
					{#if isSharedDatabase}
						<span class="shared-badge">{databasePermission} access</span>
					{/if}
				</div>
				<div class="database-actions">
					{#if canAdmin()}
						<button class="btn btn-share" on:click={() => {
							console.log('Share button clicked, showShareModal:', showShareModal);
							showShareModal = true;
							console.log('showShareModal set to:', showShareModal);
						}} title="Share Database">
							🔗 Share
						</button>
					{/if}
					{#if isSharedDatabase}
						<button class="btn btn-danger" on:click={leaveSharedDatabase} title="Leave Shared Database">
							❌ Leave Database
						</button>
					{/if}
				</div>
			</div>
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
			
			{#if $connections.length > 0}
				<h3>Your Databases</h3>
				<div class="connections-grid">
					{#each $connections as connection}
						<div class="connection-card" on:click={() => {
							selectedConnection = connection; 
							isSharedDatabase = false;
							databasePermission = 'admin';
							loadCollections();
						}}>
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
							<div class="connection-owner">Owner</div>
						</div>
					{/each}
				</div>
			{/if}

			{#if sharedDatabases.length > 0}
				<h3>Shared with You</h3>
				<div class="connections-grid">
					{#each sharedDatabases as shared}
						<div class="connection-card shared-card" on:click={() => {
							selectedConnection = shared.database; 
							isSharedDatabase = true;
							databasePermission = shared.permission_level;
							loadCollections();
						}}>
							<div class="connection-icon">
								{#if shared.database.type === 'mongodb'}
									🍃
								{:else if shared.database.type === 'mysql'}
									🐬
								{:else if shared.database.type === 'postgresql'}
									🐘
								{:else}
									💾
								{/if}
							</div>
							<h3>{shared.database.name}</h3>
							<p>{shared.database.type}</p>
							<p class="host">{shared.database.host}:{shared.database.port}</p>
							<div class="connection-permission">{shared.permission_level} access</div>
							<div class="connection-shared">Shared by {shared.grantor.email}</div>
						</div>
					{/each}
				</div>
			{/if}

			{#if $connections.length === 0 && sharedDatabases.length === 0}
				<div class="empty-state">
					<p>No database connections available</p>
					<a href="/connections" class="btn btn-primary">Add Connection</a>
				</div>
			{/if}
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
							{#if canWrite()}
								<button class="btn btn-primary" on:click={openCreateModal}>
									➕ Add Document
								</button>
							{/if}
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
						<div class="table-wrapper">
							<div class="scroll-hint">
								<span>⟵ Scroll horizontally to see all columns ⟶</span>
							</div>
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
											{#if canWrite()}
												<th class="actions-column">Actions</th>
											{/if}
										</tr>
									</thead>
									<tbody>
										{#each documents as document, index}
											<tr>
												{#each getDocumentKeys(documents) as key}
													<td class="data-cell">{formatValue(document[key])}</td>
												{/each}
												{#if canWrite()}
													<td class="actions-cell">
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
												{/if}
											</tr>
										{/each}
								</tbody>
							</table>
						</div>
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
				<label class="form-label">Document Fields</label>
				<div class="dynamic-form">
					{#each collectionSchema as field}
						{#if !['id', 'created_at', 'updated_at'].includes((typeof field === 'string' ? field : field.name).toLowerCase())}
							{@const fieldName = typeof field === 'string' ? field : field.name}
							{@const fieldType = typeof field === 'string' ? 'text' : field.type}
							<div class="form-field">
								<label class="field-label">{fieldName}</label>
								
								{#if fieldType === 'image'}
									<div class="image-input-container">
										<input
											type="file"
											accept="image/*"
											on:change={(e) => handleImageUpload(e, fieldName)}
											class="file-input"
											id="file-{fieldName}"
										/>
										<label for="file-{fieldName}" class="file-input-label">
											📷 Choose Image
										</label>
										{#if newDocumentData[fieldName]}
											<div class="image-preview">
												{#if newDocumentData[fieldName].startsWith('data:')}
													<img src={newDocumentData[fieldName]} alt="Preview" class="preview-image" />
												{:else}
													<span class="file-name">{newDocumentData[fieldName]}</span>
												{/if}
												<button type="button" class="remove-image" on:click={() => removeImage(fieldName)}>×</button>
											</div>
										{/if}
										<small class="field-hint">Upload an image file or enter image URL</small>
										<Input
											type="url"
											bind:value={newDocumentData[fieldName]}
											placeholder="Or enter image URL"
											class="form-input image-url-input"
										/>
									</div>
								{:else if fieldType === 'textarea'}
									<textarea
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter {fieldName}"
										class="form-textarea"
										rows="3"
									></textarea>
								{:else if fieldType === 'checkbox'}
									<label class="checkbox-container">
										<input
											type="checkbox"
											bind:checked={newDocumentData[fieldName]}
											class="form-checkbox"
										/>
										<span class="checkmark"></span>
										Enable {fieldName}
									</label>
								{:else if fieldType === 'number'}
									<Input
										type="number"
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter {fieldName}"
										class="form-input"
									/>
								{:else if fieldType === 'email'}
									<Input
										type="email"
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter email address"
										class="form-input"
									/>
								{:else if fieldType === 'url'}
									<Input
										type="url"
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter URL"
										class="form-input"
									/>
								{:else if fieldType === 'tel'}
									<Input
										type="tel"
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter phone number"
										class="form-input"
									/>
								{:else if fieldType === 'password'}
									<Input
										type="password"
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter password"
										class="form-input"
									/>
								{:else if fieldType === 'datetime-local'}
									<Input
										type="datetime-local"
										bind:value={newDocumentData[fieldName]}
										class="form-input"
									/>
								{:else}
									<Input
										type="text"
										bind:value={newDocumentData[fieldName]}
										placeholder="Enter {fieldName}"
										class="form-input"
									/>
								{/if}
							</div>
						{/if}
					{/each}
					
					{#if collectionSchema.length === 0}
						<p class="no-fields">Loading collection schema...</p>
					{/if}
				</div>
				<small class="form-help">
					{#if collectionSchema && collectionSchema.length > 0}
						Fill the fields for collection: <strong>{selectedCollection}</strong><br>
						Auto-generated fields (id, created_at, updated_at) are handled automatically
					{:else}
						Loading available fields...
					{/if}
				</small>
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
		max-width: 100vw;
		width: 100%;
		margin: 0 auto;
		padding: 10px; /* Reduced padding */
		box-sizing: border-box;
		overflow-x: hidden; /* Prevent horizontal overflow */
	}

	/* Responsive container */
	@media (min-width: 1400px) {
		.container {
			max-width: 1500px; /* Increased max width */
			padding: 15px; /* Slightly more padding for larger screens */
		}
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

	/* Shared Cards Styling */
	.shared-card {
		border-color: #e0f2fe;
		background: linear-gradient(135deg, #f0f9ff, #ffffff);
	}

	.shared-card:hover {
		border-color: #0ea5e9;
	}

	.connection-owner {
		background: #dcfce7;
		color: #166534;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		margin-top: 0.5rem;
		display: inline-block;
	}

	.connection-permission {
		background: #fef3c7;
		color: #92400e;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		margin-top: 0.5rem;
		display: inline-block;
		text-transform: capitalize;
	}

	.connection-shared {
		font-size: 0.75rem;
		color: #64748b;
		margin-top: 0.5rem;
		font-style: italic;
	}

	.no-connection h3 {
		color: #334155;
		font-size: 1.25rem;
		margin: 2rem 0 1rem 0;
		border-bottom: 2px solid #e2e8f0;
		padding-bottom: 0.5rem;
	}

	.empty-state {
		margin-top: 2rem;
		padding: 2rem;
		background: #f8fafc;
		border-radius: 8px;
		border: 1px dashed #cbd5e1;
	}

	.empty-state p {
		margin-bottom: 1rem;
		color: #64748b;
	}

	/* Database Management Layout */
	.database-management {
		display: grid;
		grid-template-columns: 280px 1fr;
		gap: 1rem;
		min-height: 600px;
		margin: 0 -0.5rem; /* Negative margin to use more space */
		padding: 0 0.5rem;
	}

	/* Responsive adjustments */
	@media (max-width: 1200px) {
		.database-management {
			grid-template-columns: 260px 1fr;
			gap: 0.75rem;
		}
	}

	@media (max-width: 768px) {
		.database-management {
			grid-template-columns: 1fr;
			gap: 1rem;
		}
		
		.sidebar {
			order: 2;
			max-height: 300px;
		}
		
		.main-content {
			order: 1;
		}
	}

	/* Sidebar */
	.sidebar {
		background: white;
		border-radius: 12px;
		padding: 1rem; /* Reduced padding */
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
		border: 2px solid #f1f5f9;
		min-width: 260px; /* Ensure minimum width */
		max-width: 280px; /* Maximum width */
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
		padding: 1rem; /* Reduced padding */
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
		border: 2px solid #f1f5f9;
		min-width: 0; /* Allow content to shrink */
		overflow: hidden; /* Prevent overflow */
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
	.table-wrapper {
		background: white;
		border-radius: 8px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		overflow: hidden;
		margin-bottom: 1.5rem;
		position: relative; /* For absolute positioning of scroll hint */
	}

	.scroll-hint {
		background: #f8f9fa;
		padding: 0.5rem 1rem;
		text-align: center;
		font-size: 0.875rem;
		color: #6b7280;
		border-bottom: 1px solid #e5e7eb;
		font-style: italic;
	}

	.table-container {
		overflow-x: auto;
		overflow-y: visible;
		max-width: 100%;
		width: 100%;
		-webkit-overflow-scrolling: touch; /* Smooth scrolling on mobile */
		box-sizing: border-box;
		border-radius: 8px;
		border: 1px solid #e5e7eb;
		position: relative;
		/* Enhanced scrollbar styling */
		scrollbar-width: thin;
		scrollbar-color: #cbd5e0 #f7fafc;
	}

	.table-container::-webkit-scrollbar {
		height: 12px;
	}

	.table-container::-webkit-scrollbar-track {
		background: #f7fafc;
		border-radius: 6px;
	}

	.table-container::-webkit-scrollbar-thumb {
		background: #cbd5e0;
		border-radius: 6px;
	}

	.table-container::-webkit-scrollbar-thumb:hover {
		background: #a0aec0;
	}

	.documents-table {
		width: 100%;
		min-width: 1000px; /* Force minimum width to trigger horizontal scroll */
		border-collapse: collapse;
		background: white;
		table-layout: auto; /* Auto layout for dynamic content */
	}

	.documents-table th,
	.documents-table td {
		padding: 0.5rem 0.75rem; /* Reduced padding to fit more content */
		text-align: left;
		border-bottom: 1px solid #e2e8f0;
		border-right: 1px solid #f1f5f9;
		vertical-align: top;
		min-width: 100px; /* Smaller minimum width */
		max-width: 200px; /* Reasonable max width */
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		box-sizing: border-box;
	}

	.documents-table th:last-child,
	.documents-table td:last-child {
		border-right: none; /* Remove border from last column */
	}

	/* Specific column widths for better spacing */
	.documents-table th:first-child,
	.documents-table td:first-child {
		min-width: 100px; /* Smaller for index/id columns */
		max-width: 150px;
	}

	.documents-table th:last-child,
	.documents-table td:last-child {
		min-width: 160px; /* Wider for action buttons */
		max-width: none; /* No max width restriction for actions */
		white-space: normal; /* Allow wrapping for action buttons */
	}

	/* Fixed columns for additional data */
	.fixed-column {
		min-width: 120px !important;
		max-width: 180px !important;
		white-space: nowrap;
	}

	/* Status badge styling */
	.status-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 1rem;
		font-size: 0.75rem;
		font-weight: 600;
		background: #10b981;
		color: white;
		text-align: center;
		white-space: nowrap;
	}

	.documents-table th {
		background: #f8f9fa;
		font-weight: 600;
		color: #4a5568;
		position: sticky; /* Make headers sticky */
		top: 0;
		z-index: 10;
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

	/* Remove duplicate styling - already handled above */

	.data-cell {
		min-width: 120px;
		max-width: 250px; /* Increased max width */
		padding: 0.75rem 1.25rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		cursor: pointer;
		position: relative;
		transition: background-color 0.2s ease;
	}

	.data-cell:hover {
		background-color: #f1f5f9;
		overflow: visible;
		white-space: normal;
		word-break: break-word;
		z-index: 5;
	}

	/* Action buttons styling */
	.documents-table .action-buttons {
		display: flex;
		gap: 0.5rem;
		min-width: 140px;
		justify-content: flex-end;
	}

	.documents-table .action-buttons .btn {
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
	}

	.data-cell:hover {
		background-color: #f7fafc;
		overflow: visible;
		white-space: normal;
		word-wrap: break-word;
		position: relative;
		z-index: 10;
	}

	.actions-column {
		width: 160px !important;
		min-width: 160px !important;
		max-width: 160px !important;
		text-align: center;
		position: sticky;
		right: 0;
		background: #f8f9fa !important;
		border-left: 2px solid #e2e8f0;
		z-index: 15;
		box-shadow: -3px 0 6px rgba(0,0,0,0.1);
		white-space: nowrap;
	}

	.actions-cell {
		width: 160px !important;
		min-width: 160px !important;
		max-width: 160px !important;
		text-align: center;
		position: sticky;
		right: 0;
		background: white !important;
		border-left: 2px solid #e2e8f0;
		z-index: 15;
		box-shadow: -3px 0 6px rgba(0,0,0,0.1);
		white-space: nowrap;
	}

	.action-buttons {
		display: flex;
		gap: 0.5rem;
		justify-content: center;
		align-items: center;
		flex-wrap: nowrap;
		padding: 0.25rem;
	}

	.action-buttons .btn {
		font-size: 0.75rem;
		padding: 0.25rem 0.5rem;
		white-space: nowrap;
		border-radius: 4px;
		border: 1px solid transparent;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.action-buttons .btn:hover {
		transform: translateY(-1px);
		box-shadow: 0 2px 4px rgba(0,0,0,0.1);
	}

	.btn-info {
		background-color: #17a2b8;
		color: white;
		border-color: #17a2b8;
	}

	.btn-info:hover {
		background-color: #138496;
		border-color: #117a8b;
	}

	.btn-danger {
		background-color: #dc3545;
		color: white;
		border-color: #dc3545;
	}

	.btn-danger:hover {
		background-color: #c82333;
		border-color: #bd2130;
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

	/* Dynamic Form Styles */
	.dynamic-form {
		display: grid;
		gap: 1rem;
		max-height: 400px;
		overflow-y: auto;
		padding: 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		background: #f9fafb;
	}

	.form-field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.field-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		text-transform: capitalize;
	}

	.no-fields {
		text-align: center;
		color: #6b7280;
		font-style: italic;
		margin: 2rem 0;
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

		/* Table Responsive */
		.table-wrapper {
			margin: 0 -0.5rem;
			border-radius: 0;
		}

		.table-container {
			-webkit-overflow-scrolling: touch;
			border-radius: 0;
		}

		.documents-table {
			min-width: 700px;
			font-size: 0.9rem;
		}

		.documents-table th,
		.documents-table td {
			padding: 0.5rem 0.4rem;
		}

		.data-cell {
			min-width: 80px;
			max-width: 120px;
			font-size: 0.85rem;
		}

		.actions-column,
		.actions-cell {
			width: 120px;
			min-width: 120px;
		}

		.action-buttons {
			flex-direction: column;
			gap: 0.2rem;
		}

		.action-buttons .btn {
			font-size: 0.7rem;
			padding: 0.2rem 0.3rem;
		}
	}

	@media (max-width: 480px) {
		.documents-table {
			min-width: 600px;
			font-size: 0.8rem;
		}

		.documents-table th,
		.documents-table td {
			padding: 0.4rem 0.3rem;
		}

		.data-cell {
			min-width: 70px;
			max-width: 100px;
			font-size: 0.75rem;
		}

		.actions-column,
		.actions-cell {
			width: 100px;
			min-width: 100px;
		}

		.action-buttons .btn {
			font-size: 0.65rem;
			padding: 0.15rem 0.25rem;
		}
	}

	/* Custom Scrollbar Styling */
	.table-container::-webkit-scrollbar {
		height: 8px;
		background-color: #f1f5f9;
	}

	.table-container::-webkit-scrollbar-track {
		background-color: #f1f5f9;
		border-radius: 4px;
	}

	.table-container::-webkit-scrollbar-thumb {
		background: linear-gradient(135deg, #667eea, #764ba2);
		border-radius: 4px;
		border: 1px solid #e2e8f0;
	}

	.table-container::-webkit-scrollbar-thumb:hover {
		background: linear-gradient(135deg, #5a67d8, #6b46c1);
	}

	/* Firefox Scrollbar */
	.table-container {
		scrollbar-width: thin;
		scrollbar-color: #667eea #f1f5f9;
	}

	/* Table hover effects for better UX */
	.documents-table tbody tr:hover {
		background-color: rgba(102, 126, 234, 0.03);
	}

	.documents-table tbody tr:hover td {
		background-color: transparent;
	}

	/* Sticky column indicators */
	.table-wrapper::after {
		content: '← Scroll horizontally to view more columns →';
		position: absolute;
		bottom: -30px;
		left: 50%;
		transform: translateX(-50%);
		font-size: 0.75rem;
		color: #6b7280;
		text-align: center;
		white-space: nowrap;
		opacity: 0.7;
		animation: fadeInOut 3s ease-in-out infinite;
		pointer-events: none;
	}

	@keyframes fadeInOut {
		0%, 100% { opacity: 0.3; }
		50% { opacity: 0.8; }
	}

	/* Hide scroll hint on small screens where it's obvious */
	@media (max-width: 768px) {
		.table-wrapper::after {
			display: none;
		}
	}

	/* Image Input Styling */
	.image-input-container {
		position: relative;
		border: 2px dashed #e0e7ff;
		border-radius: 8px;
		padding: 1rem;
		background: #f8faff;
		transition: all 0.2s;
	}

	.image-input-container:hover {
		border-color: #667eea;
		background: #f0f4ff;
	}

	.file-input {
		position: absolute;
		opacity: 0;
		width: 100%;
		height: 100%;
		cursor: pointer;
	}

	.file-input-label {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 500;
		font-size: 0.9rem;
		transition: all 0.2s;
		margin-bottom: 0.5rem;
	}

	.file-input-label:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
	}

	.image-preview {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0.5rem 0;
		padding: 0.5rem;
		background: white;
		border-radius: 6px;
		border: 1px solid #e0e7ff;
	}

	.preview-image {
		width: 60px;
		height: 60px;
		object-fit: cover;
		border-radius: 4px;
		border: 1px solid #ddd;
	}

	.file-name {
		font-size: 0.9rem;
		color: #555;
		max-width: 200px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.remove-image {
		background: #dc3545;
		color: white;
		border: none;
		border-radius: 50%;
		width: 24px;
		height: 24px;
		cursor: pointer;
		font-size: 14px;
		line-height: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-left: auto;
	}

	.remove-image:hover {
		background: #c82333;
	}

	.image-url-input {
		margin-top: 0.5rem;
	}

	.field-hint {
		display: block;
		margin-top: 0.25rem;
		color: #666;
		font-size: 0.8rem;
	}

	/* Checkbox Styling */
	.checkbox-container {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
		padding: 0.5rem 0;
		font-size: 0.9rem;
		color: #555;
	}

	.form-checkbox {
		width: 18px;
		height: 18px;
		cursor: pointer;
	}

	.checkmark {
		position: relative;
		display: inline-block;
		width: 20px;
		height: 20px;
		background: white;
		border: 2px solid #e0e7ff;
		border-radius: 4px;
		transition: all 0.2s;
	}

	.checkbox-container input:checked + .checkmark {
		background: linear-gradient(135deg, #667eea, #764ba2);
		border-color: #667eea;
	}

	.checkbox-container input:checked + .checkmark::after {
		content: '✓';
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		color: white;
		font-size: 12px;
		font-weight: bold;
	}

	/* Image Input Styling */
	.image-input-container {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.file-input {
		display: none;
	}

	.file-input-label {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 10px 16px;
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		font-size: 14px;
		font-weight: 500;
		transition: all 0.2s;
		max-width: 200px;
	}

	.file-input-label:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
	}

	.image-preview {
		position: relative;
		display: inline-block;
		margin-top: 8px;
	}

	.preview-image {
		max-width: 200px;
		max-height: 150px;
		border-radius: 8px;
		border: 2px solid #e0e7ff;
		object-fit: cover;
	}

	.remove-image {
		position: absolute;
		top: -8px;
		right: -8px;
		width: 24px;
		height: 24px;
		background: #ef4444;
		color: white;
		border: none;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: bold;
		transition: all 0.2s;
	}

	.remove-image:hover {
		background: #dc2626;
		transform: scale(1.1);
	}

	.image-url-input {
		margin-top: 8px;
	}

	.field-hint {
		font-size: 12px;
		color: #6b7280;
		margin-top: 4px;
	}

	.file-name {
		display: inline-block;
		padding: 8px 12px;
		background: #f3f4f6;
		border-radius: 6px;
		font-size: 13px;
		color: #374151;
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Form Textarea */
	.form-textarea {
		width: 100%;
		padding: 12px;
		border: 2px solid #e0e7ff;
		border-radius: 8px;
		font-size: 14px;
		resize: vertical;
		min-height: 80px;
		font-family: inherit;
		transition: border-color 0.2s;
	}

	.form-textarea:focus {
		outline: none;
		border-color: #667eea;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	/* Loading and Error States */
	.loading-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(255, 255, 255, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		z-index: 10;
	}

	.loading-spinner {
		width: 24px;
		height: 24px;
		border: 2px solid #e0e7ff;
		border-top: 2px solid #667eea;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	/* Responsive adjustments for forms */
	@media (max-width: 768px) {
		.preview-image {
			max-width: 150px;
			max-height: 100px;
		}
		
		.file-input-label {
			max-width: 150px;
			font-size: 13px;
			padding: 8px 12px;
		}
	}

	.btn-share {
		background: linear-gradient(135deg, #667eea, #764ba2);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		margin-left: 1rem;
		transition: all 0.2s ease;
	}

	.btn-share:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
	}

	.btn-danger {
		background: linear-gradient(135deg, #ef4444, #dc2626);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		margin-left: 1rem;
		transition: all 0.2s ease;
	}

	.btn-danger:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
		background: linear-gradient(135deg, #dc2626, #b91c1c);
	}

	.subtitle {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.subtitle-info {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.database-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.shared-badge {
		background: linear-gradient(135deg, #10b981, #059669);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		margin-left: 0.5rem;
		text-transform: capitalize;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
	}

	.shared-badge::before {
		content: "👥";
		font-size: 0.875rem;
	}
</style>

<!-- Share Database Modal -->
<ShareDatabaseModal 
	bind:isOpen={showShareModal} 
	database={selectedConnection}
	on:close={() => showShareModal = false}
/>
