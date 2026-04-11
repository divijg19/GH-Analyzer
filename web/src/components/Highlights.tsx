type HighlightsProps = {
	highlights: string[];
};

export default function Highlights(props: HighlightsProps) {
	return (
		<section class="card-section">
			<h3 class="section-title">Highlights</h3>
			{props.highlights.length > 0 ? (
				<ul class="section-list">
					{props.highlights.map((item) => (
						<li>{item}</li>
					))}
				</ul>
			) : (
				<p class="section-text">No highlights.</p>
			)}
		</section>
	);
}
