<script lang="ts">
	import type { ReadingChallenge } from '$lib/types';
	import { enhance } from '$app/forms';
	import { page } from '$app/stores';
	import { invalidate } from '$app/navigation'; 

	export let challenges: ReadingChallenge[];

	// Derive a flag for whether we're on the /challenges route.
	$: isChallengesPage = $page.url.pathname === '/reading-challenges';

	// Helper function to round to the nearest quarter
	function roundToQuarter(num: number): number {
		return Math.round(num * 4) / 4;
	}

	// If the absolute value is very small (< 0.01), we return "0.00".
	function formatRate(rate: number | undefined, showExact = false, type?: string): string {
		if (rate === undefined || rate === null) {
			return "0.00";
		}
		// If the value is extremely small, just return "0.00"
		if (Math.abs(rate) < 0.24) {
			return "0.00";
		}
		if (showExact) {
			return rate.toFixed(2);
		}
		if (type?.toLowerCase() === 'pages') {
			return Math.floor(rate).toString();
		}
		return roundToQuarter(rate).toFixed(2);
	}
</script>

<div class="flex w-full max-w-7xl flex-col items-center">
	{#each challenges as challenge}
		<div class="mb-16 w-full text-left">
			<div class="mb-4 flex items-start justify-between">
				<div>
					<h1 class="text-2xl font-bold text-gray-600 md:text-3xl lg:text-4xl">
						{challenge.name}
					</h1>
					<p class="mt-2 text-gray-600">
						Read {challenge.target} {challenge.type.toLowerCase()} by
						{new Date(challenge.endDate).toLocaleDateString()}.
					</p>
				</div>
				<!-- Only display the delete button if we're on the /challenges page -->
                {#if isChallengesPage}
                    <form
                        method="POST"
                        action="?/delete"
                        use:enhance={async () => {
                            await invalidate('/reading-challenges');
                        }}
                    >
                        <input type="hidden" name="id" value={challenge.id} />
                        <button
                            type="submit"
                            class="rounded-full bg-red-500 px-4 py-2 text-sm text-white hover:bg-red-600"
                        >
                            Delete Challenge
                        </button>
                    </form>
                {/if}
			</div>

			<div class="mt-6 w-full">
				<!-- Progress Bar Container -->
				<div class="h-8 w-full overflow-hidden rounded-full border-2 border-gray-300 bg-gray-200">
					<div
						class="h-full bg-blue-500 transition-all duration-300"
						style="width: {challenge.progress.percentage}%;"
					></div>
				</div>

				<!-- Stats Below Progress Bar -->
				<div class="mt-2 flex items-center justify-between text-sm text-gray-600">
					<span>{challenge.progress.current} {challenge.type.toLowerCase()} read</span>
					<span>
						{challenge.target - challenge.progress.current} {challenge.type.toLowerCase()} remaining
					</span>
				</div>

				<!-- Additional Stats -->
				<div class="mt-4 grid grid-cols-4 gap-4">
					<!-- Completion Percentage -->
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						<p class="text-2xl font-bold text-blue-500">
							{challenge.progress.percentage.toFixed(1)}%
						</p>
						<p class="text-xs text-gray-600">Complete</p>
					</div>

					<!-- Current Pace -->
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						<p class="text-2xl font-bold text-blue-500">
							{formatRate(challenge.progress.rate.currentPace, (challenge.progress.rate.currentPace ?? 0) < 1, challenge.type)}
						</p>
						<p class="text-xs text-gray-600">
							Current Pace ({challenge.progress.rate.unit})
						</p>
					</div>

					<!-- Target Pace -->
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						<p class="text-2xl font-bold text-blue-500">
							{formatRate(challenge.progress.rate.required, false, challenge.type)}
						</p>
						<p class="text-xs text-gray-600">
							Target Pace ({challenge.progress.rate.unit})
						</p>
					</div>

					<!-- Schedule Status -->
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						{#if challenge.progress.rate.status === 'AHEAD'}
							<p
								class="text-2xl font-bold text-green-500"
								title="Ahead by {formatRate(challenge.progress.rate.scheduleDiff, true, challenge.type)} {challenge.progress.rate.unit}"
							>
								{formatRate(challenge.progress.rate.scheduleDiff, false, challenge.type)}
							</p>
							<p class="text-xs text-gray-600">Ahead of Schedule</p>
						{:else if challenge.progress.rate.status === 'ON_TRACK'}
							<p class="text-2xl font-bold text-blue-500">On Track</p>
							<p class="text-xs text-gray-600">Keep it up!</p>
						{:else}
							<p
								class="text-2xl font-bold text-red-500"
								title="Behind by {formatRate(challenge.progress.rate.scheduleDiff, true, challenge.type)} {challenge.progress.rate.unit}"
							>
								{formatRate(challenge.progress.rate.scheduleDiff, false, challenge.type)}
							</p>
							<p class="text-xs text-gray-600">Behind Schedule</p>
						{/if}
					</div>
				</div>

				<!-- Reading Challenge Status -->
				<div class="my-24 mt-4 flex items-center gap-4">
					{#if challenge.progress.rate.status === 'AHEAD'}
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-green-500"></div>
							<span class="text-sm text-gray-600">Ahead of Schedule</span>
						</div>
					{:else if challenge.progress.rate.status === 'ON_TRACK'}
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-blue-500"></div>
							<span class="text-sm text-gray-600">On Track</span>
						</div>
					{:else}
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-red-500"></div>
							<span class="text-sm text-gray-600">Behind Schedule</span>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/each}
</div>
