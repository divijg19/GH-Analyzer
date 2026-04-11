type HighlightsProps = {
	highlights: string[];
};

export default function Highlights(props: HighlightsProps) {
	return (
		<section>
			<h3 class="text-xs uppercase tracking-wide text-gray-400">Highlights</h3>
			{props.highlights.length > 0 ? (
				<ul class="ml-5 mt-3 list-disc space-y-1">
					{props.highlights.map((item) => (
						<li>{item}</li>
					))}
				</ul>
			) : (
				<p class="mt-3 text-gray-600">No highlights.</p>
			)}
		</section>
	);
}
