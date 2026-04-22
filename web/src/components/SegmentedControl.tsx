type Mode = "dataset" | "live";

type SegmentedControlProps = {
	mode: Mode;
	onChange: (mode: Mode) => void;
};

export default function SegmentedControl(props: SegmentedControlProps) {
	return (
		<div class="grid h-10 w-52 grid-cols-2 rounded-lg border border-slate-300 bg-slate-100 p-1">
			<button
				type="button"
				onClick={() => props.onChange("dataset")}
				class="h-8 rounded-md border border-transparent px-3 text-sm font-medium"
				classList={{
					"bg-white text-slate-900": props.mode === "dataset",
					"text-slate-600": props.mode !== "dataset",
				}}
			>
				Dataset
			</button>
			<button
				type="button"
				onClick={() => props.onChange("live")}
				class="h-8 rounded-md border border-transparent px-3 text-sm font-medium"
				classList={{
					"bg-white text-slate-900": props.mode === "live",
					"text-slate-600": props.mode !== "live",
				}}
			>
				Live
			</button>
		</div>
	);
}
