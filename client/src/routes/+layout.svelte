<script lang="ts">
	import '../app.css';
	import NavBar from '$lib/components/NavBar.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import TokenMonitor from '$lib/components/TokenMonitor.svelte';
</script>

<div class="flex min-h-screen flex-col">
	<NavBar />
	<main class="flex flex-grow flex-col">
		<slot />
	</main>
	<Footer />
	<TokenMonitor />
</div>
