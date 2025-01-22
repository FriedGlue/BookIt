<script lang="ts">
	import '../app.css';
	import NavBar from '$lib/components/NavBar.svelte';
	import Footer from '$lib/components/Footer.svelte';

	export let data: { authenticated: boolean };
</script>

<div class="min-h-screen flex flex-col">
	<NavBar authenticated={data.authenticated}/>
	<main class="flex-grow flex flex-col">
		<slot />
	</main>
	<Footer />
</div>
