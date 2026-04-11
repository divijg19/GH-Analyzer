import type { Report } from "../api/client";

type TopReposProps = {
	repos: Report["top_repos"];
};

export default function TopRepos(props: TopReposProps) {
	return (
		<section
			style={{
				padding: "16px",
				border: "1px solid #e5e7eb",
				"border-radius": "6px",
			}}
		>
			<h3 style={{ margin: "0 0 12px 0", "font-size": "18px" }}>
				Top Repositories
			</h3>
			{props.repos.length > 0 ? (
				<ul style={{ margin: "0", padding: "0 0 0 18px" }}>
					{props.repos.map((repo) => (
						<li style={{ "margin-bottom": "8px" }}>
							<span>{repo.name}</span>
							<span style={{ color: "#6b7280" }}> — {repo.size}</span>
						</li>
					))}
				</ul>
			) : (
				<p style={{ margin: "0" }}>No repositories found.</p>
			)}
		</section>
	);
}
