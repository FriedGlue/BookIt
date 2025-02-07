<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import type { DecodedToken } from '$lib/types';
    import { AuthService } from '$lib/services/authService';
    import TokenExpirationModal from './TokenExpirationModal.svelte';
    import { isAuthenticated } from '$lib/stores/authStore';
    import { jwtDecode } from 'jwt-decode';

    let showModal = false;
    let checkInterval: NodeJS.Timeout;
    let timeRemaining = '';
    const authService = new AuthService();

    function formatTimeRemaining(milliseconds: number): string {
        if (milliseconds <= 0) return 'Token expired';
        
        const minutes = Math.floor(milliseconds / (1000 * 60));
        const seconds = Math.floor((milliseconds % (1000 * 60)) / 1000);
        
        return `${minutes}m ${seconds}s`;
    }

    async function checkTokenExpiration() {
        const token = await authService.getToken();

        if (!token) {
            timeRemaining = 'No token found';
            return;
        }

        try {
            const decoded = jwtDecode<DecodedToken>(token);
            const expirationTime = decoded.exp * 1000; // Convert to milliseconds
            const now = Date.now();
            const timeUntilExpiry = expirationTime - now;

            timeRemaining = formatTimeRemaining(timeUntilExpiry);

            if (timeUntilExpiry > 0 && timeUntilExpiry < 5 * 60 * 1000) {
                showModal = true;
            }
        } catch (err) {
            console.error('Failed to decode token:', err);
            timeRemaining = 'Error decoding token';
        }
    }

    async function handleRefresh() {
        try {
            await authService.refreshToken();
            showModal = false;
            // Check token expiration immediately after refresh
            checkTokenExpiration();
        } catch (err) {
            console.error('Failed to refresh token:', err);
            isAuthenticated.set(false);
            window.location.href = '/login';
        }
    }

    onMount(() => {
        // Check token expiration every second for more accurate testing
        checkInterval = setInterval(checkTokenExpiration, 1000);
        // Initial check
        checkTokenExpiration();
    });

    onDestroy(() => {
        if (checkInterval) {
            clearInterval(checkInterval);
        }
    });
</script>

{#if import.meta.env.DEV}
    <div class="fixed bottom-4 right-4 rounded-lg bg-gray-800 p-3 text-sm text-white opacity-75">
        Token expires in: {timeRemaining}
    </div>
{/if}

<TokenExpirationModal isOpen={showModal} onRefresh={handleRefresh} /> 