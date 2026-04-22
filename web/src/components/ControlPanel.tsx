import type { SearchResult } from "../api/client";
import Shortlist from "./Shortlist";

type ControlPanelProps = {
	selectedCount: number;
	canCompare: boolean;
	onCompare: () => void;
	onClearSelection: () => void;
	canAddSelected: boolean;
	onAddSelected: () => void;
	shortlist: SearchResult[];
	onRemoveShortlist: (username: string) => void;
	onClearShortlist: () => void;
	onExportJSON: () => void;
	onExportMarkdown: () => void;
};

const buttonClass =
	"h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50";

export default function ControlPanel(props: ControlPanelProps) {
	const hasShortlist = () => props.shortlist.length > 0;

	return (
		<aside class="h-full w-72 shrink-0 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="space-y-6">
				<section class="space-y-2">
					<h2 class="text-sm font-semibold text-slate-700">Selection</h2>
					<p class="text-sm text-slate-600">{props.selectedCount} selected</p>
					<button
						type="button"
						onClick={props.onCompare}
						disabled={!props.canCompare}
						class={buttonClass}
					>
						Compare ({props.selectedCount})
					</button>
					<button
						type="button"
						onClick={props.onClearSelection}
						disabled={props.selectedCount === 0}
						class={buttonClass}
					>
						Clear
					</button>
				</section>

				<section class="space-y-2">
					<h2 class="text-sm font-semibold text-slate-700">Shortlist</h2>
					<button
						type="button"
						onClick={props.onAddSelected}
						disabled={!props.canAddSelected}
						class={buttonClass}
					>
						Add Selected
					</button>
					<button
						type="button"
						onClick={props.onClearShortlist}
						disabled={!hasShortlist()}
						class={buttonClass}
					>
						Clear Shortlist
					</button>
					<button
						type="button"
						onClick={props.onExportJSON}
						disabled={!hasShortlist()}
						class={buttonClass}
					>
						Export JSON
					</button>
					<button
						type="button"
						onClick={props.onExportMarkdown}
						disabled={!hasShortlist()}
						class={buttonClass}
					>
						Export Markdown
					</button>
				</section>

				<Shortlist
					results={props.shortlist}
					onRemove={props.onRemoveShortlist}
				/>
			</div>
		</aside>
	);
}
