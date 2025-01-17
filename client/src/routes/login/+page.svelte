<script lang="ts">
    import { AuthService } from '$lib/services/authService';
    
    const authService = new AuthService();
    let username = '';
    let password = '';
    let error = '';

    async function handleLogin() {
        try {
            await authService.login(username, password);
            // Redirect to home or dashboard
        } catch (err) {
            console.error(err);
        }
    }
</script>

<form on:submit|preventDefault={handleLogin}>
    <input type="text" bind:value={username} placeholder="Username">
    <input type="password" bind:value={password} placeholder="Password">
    <button type="submit">Login</button>
    {#if error}
        <p class="error">{error}</p>
    {/if}
</form> 