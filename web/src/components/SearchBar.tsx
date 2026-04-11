import type { JSX } from "solid-js";

type SearchBarProps = {
	username: string;
	loading: boolean;
	onUsernameChange: (value: string) => void;
	onSubmit: () => void;
};

export default function SearchBar(props: SearchBarProps) {
	const isDisabled = () => props.loading || props.username.trim().length === 0;

	const handleSubmit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (
		event,
	) => {
		event.preventDefault();
		if (isDisabled()) {
			return;
		}
		props.onSubmit();
	};

	return (
		<form onSubmit={handleSubmit} class="mx-6 flex flex-1 items-center gap-3">
			<input
				type="text"
				autofocus
				placeholder="Enter GitHub username (e.g. torvalds)"
				value={props.username}
				onInput={(event) => props.onUsernameChange(event.currentTarget.value)}
				disabled={props.loading}
				class="mx-auto w-full max-w-xl rounded-md border px-4 py-2 focus:outline-none focus:ring-2"
			/>
			<button
				type="submit"
				disabled={isDisabled()}
				class="rounded-md bg-black px-4 py-2 text-white disabled:opacity-50"
			>
				Analyze
			</button>
		</form>
	);
}
