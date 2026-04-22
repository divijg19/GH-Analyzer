import { For, Show } from "solid-js";

import type { SearchResult } from "../api/client";

type SelectionPanelProps = {
	selected: Set<string>;
	results: SearchResult[];
	onRemove: (username: string) => void;
	onClear: () => void;
	onCompare: () => void;
	onExportJSON: () => void;
	onExportMD: () => void;
};

const buttonClass =
	"h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50";

export default function SelectionPanel(props: SelectionPanelProps) {
	const selectedList = () => Array.from(props.selected);

	const visibleUsernames = () => {
		const names = new Set<string>();
		for (const result of props.results) {
			if (props.selected.has(result.username)) {
				names.add(result.username);
			}
		}

		return names;
	};

	const totalSelected = () => props.selected.size;
	const visibleCount = () => visibleUsernames().size;
	const canCompare = () => visibleCount() >= 2 && visibleCount() <= 5;
	const canExport = () => visibleCount() > 0;

	return (
		<div class="space-y-4">
			<div>
				<h2 class="text-sm font-semibold text-slate-700">Selection</h2>
				<p class="mt-1 text-sm text-slate-600">
					{totalSelected()} selected ({visibleCount()} visible)
				</p>
			</div>

			<div class="space-y-2">
				<button
					type="button"
					onClick={props.onCompare}
					disabled={!canCompare()}
					class={buttonClass}
				>
					Compare ({totalSelected()} selected, {visibleCount()} visible)
				</button>
				<button
					type="button"
					onClick={props.onExportJSON}
					disabled={!canExport()}
					class={buttonClass}
				>
					Export JSON
				</button>
				<button
					type="button"
					onClick={props.onExportMD}
					disabled={!canExport()}
					class={buttonClass}
				>
					Export Markdown
				</button>
				<button
					type="button"
					onClick={props.onClear}
					disabled={totalSelected() === 0}
					class={buttonClass}
				>
					Clear Selection
				</button>
			</div>

			<div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
				<div class="mb-2 text-sm font-semibold text-slate-700">
					Selected Users
				</div>
				<div class="max-h-72 space-y-2 overflow-y-auto pr-1">
					<For each={selectedList()}>
						{(username) => (
							<div class="flex items-center justify-between rounded-md border border-slate-200 bg-white px-2 py-2 text-sm">
								<span class="truncate text-slate-700">{username}</span>
								<button
									type="button"
									onClick={() => props.onRemove(username)}
									class="h-8 rounded-md border border-slate-300 bg-white px-2 text-xs text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
								>
									Remove
								</button>
							</div>
						)}
					</For>
					<Show when={selectedList().length === 0}>
						<div class="rounded-md border border-slate-200 bg-white px-2 py-2 text-sm text-slate-500">
							No selected users
						</div>
					</Show>
				</div>
			</div>
		</div>
	);
}
