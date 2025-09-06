<script>
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { apiClient } from '$lib/api';
    import Alert from '$lib/components/Alert.svelte';
    import Button from '$lib/components/Button.svelte';

    let invitation = null;
    let loading = true;
    let accepting = false;
    let error = '';
    let success = '';
    let isLoggedIn = false;

    $: token = $page.params.code;

    onMount(async () => {
        // Check if user is logged in
        const authToken = localStorage.getItem('token');
        isLoggedIn = !!authToken;

        await loadInvitation();
    });

    async function loadInvitation() {
        loading = true;
        error = '';

        try {
            invitation = await apiClient.getInvitation(token);
        } catch (err) {
            error = err.response?.data?.error || 'Invalid or expired invitation';
        } finally {
            loading = false;
        }
    }

    async function acceptInvitation() {
        if (!isLoggedIn) {
            // Store the invitation token and redirect to login
            localStorage.setItem('pendingInvitation', token);
            goto('/login?redirect=' + encodeURIComponent('/join/' + token));
            return;
        }

        accepting = true;
        error = '';
        success = '';

        try {
            await apiClient.acceptInvitation(token);
            success = 'Invitation accepted successfully! You now have access to the database.';
            
            // Redirect to dashboard after a short delay
            setTimeout(() => {
                goto('/dashboard');
            }, 2000);
        } catch (err) {
            error = err.response?.data?.error || 'Failed to accept invitation';
        } finally {
            accepting = false;
        }
    }

    function getPermissionDescription(level) {
        switch (level) {
            case 'read':
                return 'You will have read-only access to view database structure and data.';
            case 'write':
                return 'You will have read and write access to view and modify data.';
            case 'admin':
                return 'You will have full administrative access including sharing permissions.';
            default:
                return 'You will have access to the database.';
        }
    }
</script>

<svelte:head>
    <title>Database Invitation</title>
</svelte:head>

<div class="container">
    <div class="invitation-card">
        {#if loading}
            <div class="loading-state">
                <div class="spinner large"></div>
                <p>Loading invitation...</p>
            </div>
        {:else if error}
            <div class="error-state">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="15" y1="9" x2="9" y2="15"/>
                    <line x1="9" y1="9" x2="15" y2="15"/>
                </svg>
                <h1>Invitation Not Found</h1>
                <p>{error}</p>
                <Button href="/login">Go to Login</Button>
            </div>
        {:else if invitation}
            <div class="invitation-content">
                <div class="invitation-header">
                    <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                        <line x1="8" y1="21" x2="16" y2="21"/>
                        <line x1="12" y1="17" x2="12" y2="21"/>
                    </svg>
                    <h1>Database Access Invitation</h1>
                </div>

                <div class="invitation-details">
                    <p class="invitation-message">
                        <strong>{invitation.inviter.email}</strong> has invited you to access the database:
                    </p>
                    
                    <div class="database-info">
                        <h2>{invitation.database.name}</h2>
                        <div class="database-meta">
                            <span class="database-type">{invitation.database.type}</span>
                            <span class="permission-level">{invitation.permission_level} access</span>
                        </div>
                    </div>

                    <div class="permission-description">
                        <p>{getPermissionDescription(invitation.permission_level)}</p>
                    </div>
                </div>

                {#if error}
                    <Alert type="error" message={error} />
                {/if}

                {#if success}
                    <Alert type="success" message={success} />
                {/if}

                <div class="invitation-actions">
                    {#if !isLoggedIn}
                        <p class="login-message">You need to be logged in to accept this invitation.</p>
                        <div class="action-buttons">
                            <Button href="/login?redirect={encodeURIComponent('/join/' + token)}" variant="primary">
                                Login to Accept
                            </Button>
                            <Button href="/register?redirect={encodeURIComponent('/join/' + token)}" variant="secondary">
                                Create Account
                            </Button>
                        </div>
                    {:else}
                        <div class="action-buttons">
                            <Button 
                                on:click={acceptInvitation} 
                                variant="primary" 
                                disabled={accepting}
                                loading={accepting}
                            >
                                {accepting ? 'Accepting...' : 'Accept Invitation'}
                            </Button>
                            <Button href="/dashboard" variant="secondary">
                                Decline
                            </Button>
                        </div>
                    {/if}
                </div>

                <div class="invitation-footer">
                    <p class="expires-note">
                        This invitation expires on {new Date(invitation.expires_at).toLocaleDateString()}
                    </p>
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .container {
        min-height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        padding: 1rem;
    }

    .invitation-card {
        background: white;
        border-radius: 16px;
        box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
        width: 100%;
        max-width: 500px;
        overflow: hidden;
    }

    .loading-state,
    .error-state {
        padding: 3rem 2rem;
        text-align: center;
    }

    .loading-state p,
    .error-state p {
        margin: 1rem 0 0 0;
        color: #6b7280;
    }

    .error-state svg {
        color: #ef4444;
        margin-bottom: 1rem;
    }

    .error-state h1 {
        margin: 0 0 0.5rem 0;
        color: #1f2937;
        font-size: 1.5rem;
    }

    .invitation-content {
        padding: 2rem;
    }

    .invitation-header {
        text-align: center;
        margin-bottom: 2rem;
    }

    .invitation-header svg {
        color: #667eea;
        margin-bottom: 1rem;
    }

    .invitation-header h1 {
        margin: 0;
        font-size: 1.75rem;
        font-weight: 700;
        color: #1f2937;
    }

    .invitation-details {
        margin-bottom: 2rem;
    }

    .invitation-message {
        text-align: center;
        margin-bottom: 1.5rem;
        color: #6b7280;
        line-height: 1.6;
    }

    .database-info {
        background: #f8fafc;
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        padding: 1.5rem;
        text-align: center;
        margin-bottom: 1rem;
    }

    .database-info h2 {
        margin: 0 0 0.5rem 0;
        font-size: 1.5rem;
        font-weight: 600;
        color: #1e293b;
    }

    .database-meta {
        display: flex;
        justify-content: center;
        gap: 1rem;
        flex-wrap: wrap;
    }

    .database-type,
    .permission-level {
        background: #e0e7ff;
        color: #3730a3;
        padding: 0.375rem 0.75rem;
        border-radius: 6px;
        font-size: 0.875rem;
        font-weight: 500;
        text-transform: capitalize;
    }

    .permission-level {
        background: #d1fae5;
        color: #065f46;
    }

    .permission-description {
        background: #fef3c7;
        border: 1px solid #f59e0b;
        border-radius: 8px;
        padding: 1rem;
        text-align: center;
    }

    .permission-description p {
        margin: 0;
        color: #92400e;
        font-size: 0.875rem;
        line-height: 1.5;
    }

    .invitation-actions {
        margin-bottom: 1.5rem;
    }

    .login-message {
        text-align: center;
        margin-bottom: 1rem;
        color: #6b7280;
        font-size: 0.875rem;
    }

    .action-buttons {
        display: flex;
        gap: 0.75rem;
        flex-direction: column;
    }

    .invitation-footer {
        text-align: center;
        border-top: 1px solid #e5e7eb;
        padding-top: 1rem;
    }

    .expires-note {
        margin: 0;
        color: #9ca3af;
        font-size: 0.875rem;
    }

    .spinner {
        width: 40px;
        height: 40px;
        border: 3px solid #f3f4f6;
        border-top: 3px solid #667eea;
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin: 0 auto;
    }

    .spinner.large {
        width: 48px;
        height: 48px;
        border-width: 4px;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    @media (min-width: 640px) {
        .action-buttons {
            flex-direction: row;
        }
    }
</style>
