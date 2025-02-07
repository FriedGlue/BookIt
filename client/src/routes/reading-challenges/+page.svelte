<script lang="ts">
	import type { PageData } from './$types';
	import { enhance } from '$app/forms';
	import ReadingChallenges from '$lib/components/ReadingChallenges.svelte';

	export let data: PageData;

	let showCreateForm = false;
	let challengeName = '';
	let challengeType: 'BOOKS' | 'PAGES' = 'BOOKS';
	let timeframe: 'YEAR' | 'MONTH' | 'WEEK' = 'YEAR';
	let target = '';
	let startDate = new Date().toISOString().split('T')[0];
	let endDate = '';

	$: {
		// Set a default end date based on timeframe
		const start = new Date(startDate);
		switch (timeframe) {
			case 'YEAR':
				endDate = new Date(start.getFullYear() + 1, start.getMonth(), start.getDate())
					.toISOString()
					.split('T')[0];
				break;
			case 'MONTH':
				endDate = new Date(start.getFullYear(), start.getMonth() + 1, start.getDate())
					.toISOString()
					.split('T')[0];
				break;
			case 'WEEK':
				endDate = new Date(start.getTime() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
				break;
		}
	}

	function resetForm() {
		showCreateForm = false;
		challengeName = '';
		target = '';
	}
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-8 flex items-center justify-between">
		<h1 class="text-3xl font-bold text-gray-800">Reading Challenges</h1>
		<button
			class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
			on:click={() => (showCreateForm = true)}
		>
			Create Challenge
		</button>
	</div>

	{#if data.profile?.challenges && data.profile.challenges.length > 0}
		<ReadingChallenges challenges={data.profile.challenges} />

		<div class="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
			{#each data.profile.challenges as challenge}
				<div class="rounded-lg bg-white p-6 shadow-lg">
					<h2 class="mb-4 text-xl font-semibold">{challenge.name}</h2>
					<p class="text-gray-600">
						{challenge.target}
						{challenge.type.toLowerCase()} by {new Date(challenge.endDate).toLocaleDateString()}
					</p>
					<form
						method="POST"
						action="?/delete"
						use:enhance={() => {
							return async ({ update }) => {
								await update();
							};
						}}
					>
						<input type="hidden" name="id" value={challenge.id} />
						<button
							type="submit"
							class="mt-4 w-full rounded-full bg-red-500 px-4 py-2 text-white hover:bg-red-600"
						>
							Delete Challenge
						</button>
					</form>
				</div>
			{/each}
		</div>
	{:else}
		<div class="flex w-full items-center justify-center py-16">
			<p class="text-4xl text-gray-500">No Reading Challenges Set</p>
		</div>
	{/if}
</div>

<!-- Create Challenge Modal -->
{#if showCreateForm}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
		<div class="w-full max-w-md rounded-lg bg-white p-8">
			<h2 class="mb-6 text-2xl font-bold">Create Reading Challenge</h2>

			<form
				method="POST"
				action="?/create"
				use:enhance={() => {
					return async ({ update }) => {
						await update();
						resetForm();
					};
				}}
				class="space-y-4"
			>
				<div>
					<label for="name" class="mb-2 block text-sm font-medium text-gray-700"
						>Challenge Name</label
					>
					<input
						type="text"
						id="name"
						name="name"
						bind:value={challengeName}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>

				<div>
					<label for="type" class="mb-2 block text-sm font-medium text-gray-700"
						>Challenge Type</label
					>
					<select
						id="type"
						name="type"
						bind:value={challengeType}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
					>
						<option value="BOOKS">Books</option>
						<option value="PAGES">Pages</option>
					</select>
				</div>

				<div>
					<label for="timeframe" class="mb-2 block text-sm font-medium text-gray-700"
						>Timeframe</label
					>
					<select
						id="timeframe"
						name="timeframe"
						bind:value={timeframe}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
					>
						<option value="YEAR">Year</option>
						<option value="MONTH">Month</option>
						<option value="WEEK">Week</option>
					</select>
				</div>

				<div>
					<label for="startDate" class="mb-2 block text-sm font-medium text-gray-700"
						>Start Date</label
					>
					<input
						type="date"
						id="startDate"
						name="startDate"
						bind:value={startDate}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>

				<input type="hidden" name="endDate" value={endDate} />

				<div>
					<label for="target" class="mb-2 block text-sm font-medium text-gray-700">
						Target ({challengeType === 'PAGES' ? 'Pages' : 'Books'})
					</label>
					<input
						type="number"
						id="target"
						name="target"
						bind:value={target}
						min="1"
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>

				<div class="mt-6 flex justify-end space-x-4">
					<button
						type="button"
						class="rounded-full bg-gray-200 px-6 py-2 text-gray-700 hover:bg-gray-300"
						on:click={resetForm}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
					>
						Create
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
