<script>
    // Sample data for books
    let books = [
        {
            title: 'To Kill a Mockingbird',
            author: 'Harper Lee',
            isbn: '9780061120084',
            progress: 70
        },
        {
            title: '1984',
            author: 'George Orwell',
            isbn: '9780451524935',
            progress: 40
        },
        {
            title: 'The Great Gatsby',
            author: 'F. Scott Fitzgerald',
            isbn: '9780743273565',
            progress: 90
        },
        // Add more books as needed
    ];
</script>

<nav class="flex flex-wrap items-center justify-between bg-blue-500 p-6">
	<div class="mr-6 flex flex-shrink-0 items-center text-white">
		<span class="text-6xl font-semibold tracking-tight">BookIt</span>
	</div>
	<div class="block w-full flex-grow lg:flex lg:w-auto lg:items-center">
		<div class="text-lg lg:flex-grow">
			<a
				href="#responsive-header"
				class="mr-4 mt-4 block text-teal-200 hover:text-white lg:mt-0 lg:inline-block"
			>
				Home
			</a>
			<a
				href="#responsive-header"
				class="mr-4 mt-4 block text-teal-200 hover:text-white lg:mt-0 lg:inline-block"
			>
				About
			</a>
		</div>
		<div>
			<!-- <a
                href="#"
                class="text >
        mr-2
        inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-sm font-semibold text-blue-500"
                >Sign In</a
            >
            <a
                href="#"
                class="inline-block rounded-full border-2 border-blue-500 bg-blue-500 px-4 py-2 text-sm font-semibold text-white"
                >Sign Up</a
            > -->
			<a
				href="#profile"
				class="mr-2 inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
				>View Profile</a
			>
		</div>
	</div>
</nav>

<section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
    <div class="w-full text-left mb-8">
        <h1 class="text-4xl md:text-5xl lg:text-6xl font-bold text-gray-800">Current Reads</h1>
    </div>
    <!-- Grid Configuration:
            - 1 column on small screens
            - 2 columns on medium screens
            - 3 columns on large screens
            - 4 columns on extra-large screens
    -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8 w-full">
        {#each books as book, i}
            <div class="flex flex-col rounded-lg overflow-hidden w-full duration-300 transform hover:scale-105 shadow-lg">
                <!-- Image Container -->
                <div class="w-full h-64 md:h-72 lg:h-64 bg-gray-300 flex items-center justify-center">
                    <img
                        src={`https://covers.openlibrary.org/b/isbn/${book.isbn}-L.jpg`}
                        alt={`Cover of ${book.title} by ${book.author}`}
                        class="max-w-full max-h-full object-contain"
                        loading="lazy"
                    />
                </div>
                <!-- Content Container -->
                <div class="flex flex-col flex-grow p-6 bg-white">
                    <h2 class="text-xl font-semibold text-gray-800 mb-2">{book.title}</h2>
                    <p class="text-gray-600 mb-4">{book.author}</p>
                    <!-- Progress Bar Container -->
                    <div class="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
                        <div
                            class="h-full bg-green-500 transition-all duration-300"
                            style="width: {book.progress}%"
                        ></div>
                    </div>
                    <!-- Display Progress Percentage and Controls -->
                    <div class="mt-2 flex justify-between items-center">
                        <span class="text-sm text-gray-700">{book.progress}%</span>
                        <div class="space-x-2">
                        </div>
                    </div>
                </div>
            </div>
        {/each}
    </div>
</section>


<!-- Example Divider -->
<hr class="my-16 border-gray-300" />

<!-- List Examples -->
<section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
    <div class="w-full text-left mb-8">
        <h1 class="text-4xl md:text-5xl lg:text-4xl font-bold text-gray-600">To Be Read (12)</h1>
        <button class="text-lg mt-2 font-semibold text-blue-500">View All</button>
    </div>
    
    <!-- Horizontal Scroll Container -->
    <div class="flex space-x-4 overflow-x-auto pb-4">
        {#each Array.from({ length: 6 }) as _, i}
            <div class="flex flex-col rounded-lg bg-gray-300 shadow-lg max-w-md hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105">
                <div class="relative w-full">
                    <img
                        src="https://covers.openlibrary.org/b/isbn/9780385472579-L.jpg"
                        alt="Cover of Book Title by Author Name"
                        class="w-full h-64 sm:h-72 md:h-80 lg:h-64 object-cover"
                    />
                </div>
            </div>
        {/each}
    </div>
</section>

<!-- List Examples -->
<section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
    <div class="w-full text-left mb-8">
        <h1 class="text-4xl md:text-5xl lg:text-4xl font-bold text-gray-600">Read (5)</h1>
        <button class="text-lg mt-2 font-semibold text-blue-500">View All</button>
    </div>
    
    <!-- Horizontal Scroll Container -->
    <div class="flex space-x-4 overflow-x-auto pb-4">
        {#each Array.from({ length: 5 }) as _, i}
            <div class="flex flex-col rounded-lg bg-gray-300 shadow-lg max-w-md hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105">
                <div class="relative w-full">
                    <img
                        src="https://covers.openlibrary.org/b/isbn/9780385472579-L.jpg"
                        alt="Cover of Book Title by Author Name"
                        class="w-full h-64 sm:h-72 md:h-80 lg:h-64 object-cover"
                    />
                </div>
            </div>
        {/each}
    </div>
</section>
