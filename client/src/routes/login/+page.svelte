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
            error = 'Invalid username or password';
        }
    }
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-200">
    <div class="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow-md">
        <div class="text-center">
            <h2 class="text-3xl font-bold text-gray-900">Sign in to your account</h2>
        </div>
        <form on:submit|preventDefault={handleLogin} class="mt-8 space-y-6">
            <div class="space-y-4">
                <div>
                    <input 
                        type="text" 
                        bind:value={username} 
                        placeholder="Username"
                        class="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10"
                    >
                </div>
                <div>
                    <input 
                        type="password" 
                        bind:value={password} 
                        placeholder="Password"
                        class="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10"
                    >
                </div>
            </div>

            <div>
                <button 
                    type="submit"
                    class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-200"
                >
                    Sign in
                </button>
            </div>
            
            {#if error}
                <p class="text-red-500 text-sm text-center mt-2">{error}</p>
            {/if}
        </form>
    </div>
</div> 