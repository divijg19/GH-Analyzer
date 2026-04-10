type HighlightsProps = {
	highlights: string[];
};

export default function Highlights(props: HighlightsProps) {
	return (
		<section
			style={{
				padding: "16px 0",
				border: "1px solid #e5e7eb",
				"border-radius": "6px",
			}}
		>
			<h3 style={{ margin: "0 16px 12px 16px", "font-size": "18px" }}>
				Highlights
			</h3>
			{props.highlights.length > 0 ? (
				<ul style={{ margin: "0", padding: "0 16px 0 32px" }}>
					{props.highlights.map((item) => (
						<li style={{ "margin-bottom": "8px" }}>{item}</li>
					))}
				</ul>
			) : (
				<p style={{ margin: "0 16px" }}>No highlights.</p>
			)}
		</section>
	);
}
