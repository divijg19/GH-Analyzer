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
				class="w-full flex-1 rounded-md border border-gray-200 px-5 py-2 text-sm transition-shadow duration-150 focus:outline-none focus:ring-2 focus:ring-black/20 disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-500"
			/>
			<button
				type="submit"
				disabled={isDisabled()}
				class="cursor-pointer rounded-md bg-black px-4 py-2.5 text-white transition-all duration-150 hover:bg-black/90 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
			>
				Analyze
			</button>
		</form>
	);
}
