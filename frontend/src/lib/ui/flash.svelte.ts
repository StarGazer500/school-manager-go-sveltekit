// A one-off success message shared by every page, e.g. "Student deleted".
// It lives in a .svelte.ts file because the page that sets it is often not
// the page that shows it: StudentDetails deletes, then the list shows the message.
// The layout clears it on every navigation, so to show a message on the page
// you are going to, call showMessage after `await goto(...)`.

export const flash = $state({ message: '' });

export function showMessage(message: string) {
	flash.message = message;
}

export function dismiss() {
	flash.message = '';
}
