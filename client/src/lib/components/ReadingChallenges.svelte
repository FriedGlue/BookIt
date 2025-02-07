<script lang="ts">
	import type { ReadingChallenge } from '$lib/types';

	export let challenges: ReadingChallenge[];
</script>

<div class="flex w-full max-w-7xl flex-col items-center">
	{#each challenges as challenge}
		<div class="mb-8 w-full text-left">
			<h1 class="text-2xl font-bold text-gray-600 md:text-3xl lg:text-4xl">{challenge.name}</h1>
			<p class="my-4 text-gray-600">
				Read {challenge.target}
				{challenge.type.toLowerCase()} by {new Date(challenge.endDate).toLocaleDateString()}.
			</p>
			<div class="mt-4 w-full">
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
					<span
						>{challenge.target - challenge.progress.current}
						{challenge.type.toLowerCase()} remaining</span
					>
				</div>

				<!-- Additional Stats -->
				<div class="mt-4 grid grid-cols-3 gap-4">
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						<p class="text-2xl font-bold text-blue-500">{challenge.progress.percentage}%</p>
						<p class="text-xs text-gray-600">Complete</p>
					</div>
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						<p class="text-2xl font-bold text-blue-500">{challenge.progress.rate.current}</p>
						<p class="text-xs text-gray-600">{challenge.progress.rate.unit}</p>
					</div>
					<div class="rounded-lg bg-gray-100 p-4 text-center">
						<p class="text-2xl font-bold text-blue-500">{challenge.progress.rate.required}</p>
						<p class="text-xs text-gray-600">{challenge.progress.rate.unit} Needed</p>
					</div>
				</div>

				<!-- Reading Challenge Status -->
				<div class="mt-6 flex items-center gap-4">
					{#if challenge.progress.rate.current >= challenge.progress.rate.required}
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-green-500"></div>
							<span class="text-sm text-gray-600">On Track</span>
						</div>
					{:else if challenge.progress.rate.current >= challenge.progress.rate.required * 0.7}
						<div class="flex items-center gap-2">
							<div class="h-3 w-3 rounded-full bg-yellow-500"></div>
							<span class="text-sm text-gray-600">Slightly Behind</span>
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
