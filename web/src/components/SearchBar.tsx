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
		<form onSubmit={handleSubmit} class="search-form">
			<input
				type="text"
				autofocus
				placeholder="Enter GitHub username (e.g. torvalds)"
				value={props.username}
				onInput={(event) => props.onUsernameChange(event.currentTarget.value)}
				disabled={props.loading}
				class="search-input"
			/>
			<button type="submit" disabled={isDisabled()} class="search-button">
				Analyze
			</button>
		</form>
	);
}
