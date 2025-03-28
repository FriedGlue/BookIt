<script lang="ts">
	import type { PageData } from './$types';
	import { enhance } from '$app/forms';
	import ReadingChallenges from '$lib/components/ReadingChallenges.svelte';
    import Sidebar from '$lib/components/Sidebar.svelte';

	export let data: PageData;

    // Helper function to round to nearest quarter
    function roundToQuarter(num: number): number {
        return Math.round(num * 4) / 4;
    }

    // Helper function to format rate for display
    function formatRate(rate: number, showExact = false): string {
        if (showExact) {
            return rate.toFixed(2);
        }
        return roundToQuarter(rate).toFixed(2);
    }

    $: options = [
        { id: 'all', label: 'All Challenges' },
        ...(data.profile?.challenges?.map(challenge => ({
            id: challenge.id,
            label: challenge.name
        })) || [])
    ];

    let selectedList = 'All Challenges';

    $: filteredChallenges = selectedList === 'All Challenges' 
        ? data.profile?.challenges || []
        : data.profile?.challenges?.filter(challenge => challenge.name === selectedList) || [];

	let showCreateForm = false;
	let challengeName = '';
	let challengeType: 'BOOKS' | 'PAGES' = 'BOOKS';
	let timeframe: 'YEAR' | 'MONTH' | 'WEEK' = 'YEAR';
	let target = '';
	let startDate = new Date().toISOString().split('T')[0];
	let endDate = '';
    let selectedSuggestion: string | null = null;

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
		selectedSuggestion = null;
	}

    // Helper function to check if a challenge type exists
    function hasExistingChallenge(type: 'YEAR' | 'MONTH' | 'WEEK'): boolean {
        return data.profile?.challenges?.some(challenge => 
            challenge.timeframe === type && 
            new Date(challenge.endDate) > new Date()  // Only consider active challenges
        ) ?? false;
    }

    function getSuggestions() {
        const suggestions = [];
        const currentDate = new Date();
        const currentYear = currentDate.getFullYear();
        const currentMonth = currentDate.toLocaleString('default', { month: 'long' });
        
        if (!hasExistingChallenge('YEAR')) {
            suggestions.push({
                name: `${currentYear} Reading Challenge`,
                type: 'BOOKS',
                timeframe: 'YEAR',
                target: '52',
                startDate: `${currentYear}-01-01`,
                description: `Read 52 books this year`
            });
        }
        
        if (!hasExistingChallenge('MONTH')) {
            suggestions.push({
                name: `${currentMonth} Reading Challenge`,
                type: 'BOOKS',
                timeframe: 'MONTH',
                target: '4',
                startDate: currentDate.toISOString().split('T')[0],
                description: `Read 4 books this month`
            });
        }
        
        if (!hasExistingChallenge('WEEK')) {
            suggestions.push({
                name: 'Weekly Page Challenge',
                type: 'PAGES',
                timeframe: 'WEEK',
                target: '500',
                startDate: currentDate.toISOString().split('T')[0],
                description: 'Read 500 pages this week'
            });
        }
        
        return suggestions;
    }

    function applySuggestion(suggestion: any) {
        selectedSuggestion = suggestion.name;
        challengeName = suggestion.name;
        challengeType = suggestion.type;
        timeframe = suggestion.timeframe;
        target = suggestion.target;
        startDate = suggestion.startDate;
    }
</script>

<div class="flex min-h-screen">
	<Sidebar
		title="View By"
		items={options.map(opt => opt.label)}
		selectedItem={selectedList}
		onSelect={(item) => (selectedList = item)}
	/>

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
            <div class="flex w-full items-center justify-center py-16">
                <ReadingChallenges challenges={filteredChallenges} />
            </div>  
        {:else}
            <div class="flex w-full items-center justify-center py-16">
                <p class="text-4xl text-gray-500">No Reading Challenges Set</p>
            </div>
        {/if}
    </div>

    <!-- Create Challenge Modal -->
    {#if showCreateForm}
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 p-4">
            <div class="w-full max-w-2xl rounded-lg bg-white p-4 sm:p-8 max-h-[90vh] overflow-y-auto">
                <h2 class="mb-4 sm:mb-6 text-xl sm:text-2xl font-bold">Create Reading Challenge</h2>

                <!-- Suggestions Section -->
                {#if getSuggestions().length > 0}
                    <div class="mb-6 sm:mb-8">
                        <h3 class="mb-3 sm:mb-4 text-base sm:text-lg font-semibold text-gray-700">Suggestions</h3>
                        <div class="grid grid-cols-1 gap-3 sm:gap-4 md:grid-cols-3">
                            {#each getSuggestions() as suggestion}
                                <button
                                    class="group relative flex flex-col rounded-lg border-2 p-3 sm:p-4 text-left transition-all duration-200 ease-in-out hover:scale-105 hover:shadow-lg
                                        {selectedSuggestion === suggestion.name 
                                            ? 'border-blue-500 bg-blue-50 shadow-md' 
                                            : 'border-blue-200 hover:border-blue-400 hover:bg-blue-50'}"
                                    on:click={() => applySuggestion(suggestion)}
                                >
                                    <div class="flex items-center justify-between">
                                        <h4 class="text-sm sm:text-base font-semibold {selectedSuggestion === suggestion.name ? 'text-blue-600' : 'text-blue-500'}">{suggestion.name}</h4>
                                        {#if selectedSuggestion === suggestion.name}
                                            <svg class="h-4 w-4 sm:h-5 sm:w-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                                                <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
                                            </svg>
                                        {/if}
                                    </div>
                                    <p class="mt-1 sm:mt-2 text-xs sm:text-sm text-gray-600">{suggestion.description}</p>
                                </button>
                            {/each}
                        </div>
                    </div>
                {/if}

                <form
                    method="POST"
                    action="?/create"
                    use:enhance={() => {
                        return async ({ update }) => {
                            await update();
                            resetForm();
                        };
                    }}
                    class="space-y-3 sm:space-y-4"
                >
                    <div>
                        <label for="name" class="mb-1 sm:mb-2 block text-xs sm:text-sm font-medium text-gray-700"
                            >Challenge Name</label
                        >
                        <input
                            type="text"
                            id="name"
                            name="name"
                            bind:value={challengeName}
                            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                            required
                        />
                    </div>

                    <div>
                        <label for="type" class="mb-1 sm:mb-2 block text-xs sm:text-sm font-medium text-gray-700"
                            >Challenge Type</label
                        >
                        <select
                            id="type"
                            name="type"
                            bind:value={challengeType}
                            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                        >
                            <option value="BOOKS">Books</option>
                            <option value="PAGES">Pages</option>
                        </select>
                    </div>

                    <div>
                        <label for="timeframe" class="mb-1 sm:mb-2 block text-xs sm:text-sm font-medium text-gray-700"
                            >Timeframe</label
                        >
                        <select
                            id="timeframe"
                            name="timeframe"
                            bind:value={timeframe}
                            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                        >
                            <option value="YEAR">Year</option>
                            <option value="MONTH">Month</option>
                            <option value="WEEK">Week</option>
                        </select>
                    </div>

                    <div>
                        <label for="startDate" class="mb-1 sm:mb-2 block text-xs sm:text-sm font-medium text-gray-700"
                            >Start Date</label
                        >
                        <input
                            type="date"
                            id="startDate"
                            name="startDate"
                            bind:value={startDate}
                            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                            required
                        />
                    </div>

                    <input type="hidden" name="endDate" value={endDate} />

                    <div>
                        <label for="target" class="mb-1 sm:mb-2 block text-xs sm:text-sm font-medium text-gray-700">
                            Target ({challengeType === 'PAGES' ? 'Pages' : 'Books'})
                        </label>
                        <input
                            type="number"
                            id="target"
                            name="target"
                            bind:value={target}
                            min="1"
                            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 text-sm"
                            required
                        />
                    </div>

                    <div class="mt-4 sm:mt-6 flex justify-end space-x-3 sm:space-x-4">
                        <button
                            type="button"
                            class="rounded-full bg-gray-200 px-4 sm:px-6 py-1.5 sm:py-2 text-xs sm:text-sm text-gray-700 hover:bg-gray-300"
                            on:click={resetForm}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            class="rounded-full bg-blue-500 px-4 sm:px-6 py-1.5 sm:py-2 text-xs sm:text-sm text-white hover:bg-blue-600"
                        >
                            Create
                        </button>
                    </div>
                </form>
            </div>
        </div>
    {/if}
</div>