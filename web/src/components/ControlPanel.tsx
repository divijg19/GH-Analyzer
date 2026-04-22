import type { SearchResult } from "../api/client";
import SelectionPanel from "./SelectionPanel";

type ControlPanelProps = {
	selected: Set<string>;
	results: SearchResult[];
	onCompare: () => void;
	onRemove: (username: string) => void;
	onClear: () => void;
	onExportJSON: () => void;
	onExportMD: () => void;
};

export default function ControlPanel(props: ControlPanelProps) {
	return (
		<aside class="h-full w-72 shrink-0 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
			<div class="h-full min-h-0 overflow-y-auto">
				<SelectionPanel
					selected={props.selected}
					results={props.results}
					onRemove={props.onRemove}
					onClear={props.onClear}
					onCompare={props.onCompare}
					onExportJSON={props.onExportJSON}
					onExportMD={props.onExportMD}
				/>
			</div>
		</aside>
	);
}
