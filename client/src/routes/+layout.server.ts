import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	// If you set event.locals.token in hooks.server.ts,
	// then you can check that here:
	return {
		authenticated: !!locals.token
	};
};
