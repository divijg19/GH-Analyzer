import type { JSX } from "solid-js";
import { createSignal } from "solid-js";

type SearchBarProps = {
	onSearch: (query: string) => void;
	live: boolean;
	onToggleLive: () => void;
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
				class="w-full min-w-0 flex-1 rounded-lg border border-slate-300 bg-white px-4 py-2.5 text-sm text-slate-800 shadow-sm outline-none transition focus:border-slate-400"
			/>
			<button
				type="submit"
				disabled={query().trim().length === 0}
				class="rounded-lg border border-slate-300 bg-slate-900 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
			>
				Search
			</button>
			<button
				type="button"
				onClick={props.onToggleLive}
				class="rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-700 shadow-sm transition hover:bg-slate-50"
			>
				{props.live ? "Live" : "Dataset"}
			</button>
		</form>
	);
}
