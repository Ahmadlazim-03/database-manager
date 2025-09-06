<script>
    import { createEventDispatcher } from 'svelte';
    import { apiClient } from '$lib/api';
    import Alert from './Alert.svelte';

    export let isOpen = false;
    export let database = null;

    let email = '';
    let permissionLevel = 'read';
    let loading = false;
    let error = '';
    let success = '';
    let invitations = [];
    let invitationLink = '';
    let loadingInvitations = false;

    const dispatch = createEventDispatcher();

    async function loadInvitations() {
        if (!database) return;
        
        loadingInvitations = true;
        try {
            invitations = await apiClient.getDatabaseInvitations(database.id);
        } catch (err) {
            console.error('Failed to load invitations:', err);
        } finally {
            loadingInvitations = false;
        }
    }

    async function handleInvite() {
        if (!email.trim()) {
            error = 'Email is required';
            return;
        }

        loading = true;
        error = '';
        success = '';
        invitationLink = '';

        try {
            const response = await apiClient.createInvitation({
                database_id: database.id,
                invitee_email: email,
                permission_level: permissionLevel
            });
            
            success = 'Invitation sent successfully!';
            invitationLink = response.invitation_link;
            email = '';
            permissionLevel = 'read';
            await loadInvitations();
        } catch (err) {
            error = err.response?.data?.error || 'Failed to send invitation';
        } finally {
            loading = false;
        }
    }

    async function handleRevokeInvitation(invitationId) {
        try {
            await apiClient.revokeInvitation(invitationId);
            success = 'Invitation revoked successfully';
            await loadInvitations();
        } catch (err) {
            error = err.response?.data?.error || 'Failed to revoke invitation';
        }
    }

    function copyInvitationLink() {
        navigator.clipboard.writeText(invitationLink);
        success = 'Invitation link copied to clipboard!';
        setTimeout(() => success = '', 3000);
    }

    function closeModal() {
        isOpen = false;
        email = '';
        error = '';
        success = '';
        invitationLink = '';
        permissionLevel = 'read';
        dispatch('close');
    }

    function formatDate(dateString) {
        return new Date(dateString).toLocaleDateString();
    }

    function getPermissionText(level) {
        switch (level) {
            case 'read': return 'Read Only';
            case 'write': return 'Read & Write';
            case 'admin': return 'Admin';
            default: return level;
        }
    }

    function getStatusBadgeClass(status) {
        switch (status) {
            case 'pending': return 'status-pending';
            case 'accepted': return 'status-accepted';
            case 'rejected': return 'status-rejected';
            case 'expired': return 'status-expired';
            default: return 'status-pending';
        }
    }

    $: if (isOpen && database) {
        loadInvitations();
    }
</script>

{#if isOpen}
    <div class="modal-overlay" on:click={closeModal}>
        <div class="modal-content" on:click|stopPropagation>
            <div class="modal-header">
                <h2>Share Database: {database?.name}</h2>
                <button class="close-btn" on:click={closeModal}>
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="18" y1="6" x2="6" y2="18"/>
                        <line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                {#if error}
                    <Alert type="error" message={error} />
                {/if}

                {#if success}
                    <Alert type="success" message={success} />
                {/if}

                <!-- Send Invitation Form -->
                <div class="section">
                    <h3>Send Invitation</h3>
                    <div class="form-group">
                        <label for="email">Email Address</label>
                        <input
                            id="email"
                            type="email"
                            bind:value={email}
                            placeholder="Enter email address"
                            disabled={loading}
                        />
                    </div>

                    <div class="form-group">
                        <label for="permission">Permission Level</label>
                        <select id="permission" bind:value={permissionLevel} disabled={loading}>
                            <option value="read">Read Only - Can view database structure and data</option>
                            <option value="write">Read & Write - Can view and modify data</option>
                            <option value="admin">Admin - Full access including sharing</option>
                        </select>
                    </div>

                    <button
                        class="btn btn-primary"
                        on:click={handleInvite}
                        disabled={loading || !email.trim()}
                    >
                        {#if loading}
                            <div class="spinner"></div>
                            Sending...
                        {:else}
                            Send Invitation
                        {/if}
                    </button>

                    {#if invitationLink}
                        <div class="invitation-link">
                            <p>Invitation link generated:</p>
                            <div class="link-container">
                                <input type="text" value={invitationLink} readonly />
                                <button class="btn btn-secondary" on:click={copyInvitationLink}>
                                    Copy
                                </button>
                            </div>
                        </div>
                    {/if}
                </div>

                <!-- Existing Invitations -->
                <div class="section">
                    <h3>Invitations</h3>
                    
                    {#if loadingInvitations}
                        <div class="loading">Loading invitations...</div>
                    {:else if invitations.length === 0}
                        <div class="empty-state">No invitations sent yet</div>
                    {:else}
                        <div class="invitations-list">
                            {#each invitations as invitation}
                                <div class="invitation-item">
                                    <div class="invitation-info">
                                        <div class="email">{invitation.invitee_email}</div>
                                        <div class="details">
                                            <span class="permission">{getPermissionText(invitation.permission_level)}</span>
                                            <span class="status {getStatusBadgeClass(invitation.status)}">
                                                {invitation.status}
                                            </span>
                                            <span class="date">Sent {formatDate(invitation.created_at)}</span>
                                        </div>
                                    </div>
                                    
                                    {#if invitation.status === 'pending'}
                                        <button
                                            class="btn btn-danger btn-sm"
                                            on:click={() => handleRevokeInvitation(invitation.id)}
                                        >
                                            Revoke
                                        </button>
                                    {/if}
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
        padding: 1rem;
    }

    .modal-content {
        background: white;
        border-radius: 12px;
        width: 100%;
        max-width: 600px;
        max-height: 90vh;
        overflow-y: auto;
        box-shadow: 0 20px 25px rgba(0, 0, 0, 0.1);
    }

    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1.5rem;
        border-bottom: 1px solid #e5e7eb;
    }

    .modal-header h2 {
        margin: 0;
        font-size: 1.5rem;
        font-weight: 600;
        color: #1f2937;
    }

    .close-btn {
        background: none;
        border: none;
        padding: 0.5rem;
        cursor: pointer;
        color: #6b7280;
        border-radius: 6px;
        transition: all 0.2s ease;
    }

    .close-btn:hover {
        background: #f3f4f6;
        color: #374151;
    }

    .modal-body {
        padding: 1.5rem;
    }

    .section {
        margin-bottom: 2rem;
    }

    .section h3 {
        margin: 0 0 1rem 0;
        font-size: 1.25rem;
        font-weight: 600;
        color: #1f2937;
    }

    .form-group {
        margin-bottom: 1rem;
    }

    .form-group label {
        display: block;
        font-size: 0.875rem;
        font-weight: 600;
        color: #374151;
        margin-bottom: 0.5rem;
    }

    .form-group input,
    .form-group select {
        width: 100%;
        padding: 0.75rem;
        border: 2px solid #e5e7eb;
        border-radius: 8px;
        font-size: 1rem;
        transition: border-color 0.2s ease;
    }

    .form-group input:focus,
    .form-group select:focus {
        outline: none;
        border-color: #667eea;
    }

    .btn {
        padding: 0.75rem 1.5rem;
        border: none;
        border-radius: 8px;
        font-size: 1rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s ease;
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
    }

    .btn-primary {
        background: linear-gradient(135deg, #667eea, #764ba2);
        color: white;
    }

    .btn-primary:hover:not(:disabled) {
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
    }

    .btn-secondary {
        background: #f3f4f6;
        color: #374151;
    }

    .btn-secondary:hover {
        background: #e5e7eb;
    }

    .btn-danger {
        background: #ef4444;
        color: white;
    }

    .btn-danger:hover {
        background: #dc2626;
    }

    .btn-sm {
        padding: 0.5rem 1rem;
        font-size: 0.875rem;
    }

    .btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .spinner {
        width: 20px;
        height: 20px;
        border: 2px solid transparent;
        border-top: 2px solid currentColor;
        border-radius: 50%;
        animation: spin 1s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .invitation-link {
        margin-top: 1rem;
        padding: 1rem;
        background: #f0f9ff;
        border: 1px solid #bfdbfe;
        border-radius: 8px;
    }

    .invitation-link p {
        margin: 0 0 0.5rem 0;
        font-size: 0.875rem;
        color: #1e40af;
    }

    .link-container {
        display: flex;
        gap: 0.5rem;
    }

    .link-container input {
        flex: 1;
        font-size: 0.875rem;
        background: white;
    }

    .invitations-list {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .invitation-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1rem;
        background: #f9fafb;
        border: 1px solid #e5e7eb;
        border-radius: 8px;
    }

    .invitation-info {
        flex: 1;
    }

    .email {
        font-weight: 600;
        color: #1f2937;
        margin-bottom: 0.25rem;
    }

    .details {
        display: flex;
        gap: 0.75rem;
        align-items: center;
        font-size: 0.875rem;
    }

    .permission {
        background: #e0e7ff;
        color: #3730a3;
        padding: 0.25rem 0.5rem;
        border-radius: 4px;
        font-size: 0.75rem;
    }

    .status {
        padding: 0.25rem 0.5rem;
        border-radius: 4px;
        font-size: 0.75rem;
        font-weight: 600;
        text-transform: capitalize;
    }

    .status-pending {
        background: #fef3c7;
        color: #92400e;
    }

    .status-accepted {
        background: #d1fae5;
        color: #065f46;
    }

    .status-rejected {
        background: #fee2e2;
        color: #991b1b;
    }

    .status-expired {
        background: #f3f4f6;
        color: #6b7280;
    }

    .date {
        color: #6b7280;
    }

    .loading {
        text-align: center;
        padding: 2rem;
        color: #6b7280;
    }

    .empty-state {
        text-align: center;
        padding: 2rem;
        color: #6b7280;
        font-style: italic;
    }
</style>
