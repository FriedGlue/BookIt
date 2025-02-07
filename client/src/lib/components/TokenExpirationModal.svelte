<script lang="ts">
    export let isOpen = false;
    export let onRefresh: () => void;
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- Backdrop -->
        <div class="fixed inset-0 bg-black bg-opacity-50"></div>
        
        <!-- Modal -->
        <div class="relative z-50 w-full max-w-md rounded-lg bg-white p-6 shadow-lg">
            <h2 class="mb-4 text-xl font-bold">Session Expired</h2>
            <p class="mb-6 text-gray-600">
                Your session has expired. Please refresh your credentials to continue.
            </p>
            <div class="flex justify-end">
                <button
                    on:click={onRefresh}
                    class="rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
                >
                    Refresh Session
                </button>
            </div>
        </div>
    </div>
{/if} 