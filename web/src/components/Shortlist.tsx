import { For } from "solid-js";

import type { SearchResult } from "../api/client";

type ShortlistProps = {
	results: SearchResult[];
	onRemove: (username: string) => void;
};

export default function Shortlist(props: ShortlistProps) {
	return (
		<div class="rounded-xl border border-slate-200 bg-slate-50 p-4">
			<div class="mb-3 flex items-center justify-between">
				<h2 class="text-sm font-semibold text-slate-700">
					Shortlist ({props.results.length})
				</h2>
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
								class="h-9 rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
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
