<script lang="ts">
	import '../app.css';
	import NavBar from '$lib/components/NavBar.svelte';
	import Footer from '$lib/components/Footer.svelte';

	export let data: { authenticated: boolean };
</script>

<div class="flex min-h-screen flex-col">
	<NavBar authenticated={data.authenticated} />
	<main class="flex flex-grow flex-col">
		<slot />
	</main>
	<Footer />
</div>
