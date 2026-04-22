import type { JSX } from "solid-js";
import { createSignal } from "solid-js";

type SearchBarProps = {
	onSearch: (query: string) => void;
	loading: boolean;
};

export default function SearchBar(props: SearchBarProps) {
	const [query, setQuery] = createSignal("");

	const handleSubmit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (
		event,
	) => {
		event.preventDefault();
		const value = query().trim();
		if (!value) {
			return;
		}

		props.onSearch(value);
	};

	return (
		<form onSubmit={handleSubmit} class="flex w-full items-center gap-3">
			<input
				type="text"
				autofocus
				placeholder="Search developers (e.g. backend or consistency > 0.7)"
				value={query()}
				onInput={(event) => setQuery(event.currentTarget.value)}
				disabled={props.loading}
				class="h-10 w-full min-w-0 flex-1 rounded-lg border border-slate-300 bg-white px-4 text-sm text-slate-800 shadow-sm outline-none focus:border-slate-400 disabled:cursor-not-allowed disabled:opacity-50"
			/>
			<button
				type="submit"
				disabled={props.loading || query().trim().length === 0}
				class="h-10 rounded-lg border border-slate-300 bg-slate-900 px-4 text-sm font-medium text-white shadow-sm hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
			>
				Search
			</button>
		</form>
	);
}
