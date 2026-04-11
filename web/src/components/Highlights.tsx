type HighlightsProps = {
	highlights: string[];
};

export default function Highlights(props: HighlightsProps) {
	return (
		<section class="mt-4">
			<h3 class="text-sm font-semibold">Highlights</h3>
			{props.highlights.length > 0 ? (
				<ul class="ml-5 mt-2 list-disc space-y-1">
					{props.highlights.map((item) => (
						<li>{item}</li>
					))}
				</ul>
			) : (
				<p class="mt-2 text-gray-600">No highlights.</p>
			)}
		</section>
	);
}
