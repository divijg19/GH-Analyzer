import { For } from "solid-js";

import type { SearchResult } from "../api/client";

type ShortlistProps = {
	results: SearchResult[];
	onRemove: (username: string) => void;
	onClear: () => void;
};

export default function Shortlist(props: ShortlistProps) {
	return (
		<div class="rounded-xl border border-slate-200 bg-slate-50 p-4">
			<div class="mb-3 flex items-center justify-between">
				<h2 class="text-sm font-semibold text-slate-700">
					Shortlist ({props.results.length})
				</h2>
				<button
					type="button"
					onClick={props.onClear}
					class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
				>
					Clear Shortlist
				</button>
			</div>

			<div class="space-y-2">
				<For each={props.results}>
					{(result) => (
						<div class="flex items-center justify-between rounded-md border border-slate-200 bg-white px-3 py-2 text-sm">
							<div class="flex items-center gap-3">
								<p class="font-medium text-slate-800">{result.username}</p>
								<p class="text-slate-600">{result.score.toFixed(2)}</p>
							</div>
							<button
								type="button"
								onClick={() => props.onRemove(result.username)}
								class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
							>
								Remove
							</button>
						</div>
					)}
				</For>
			</div>
		</div>
	);
}
