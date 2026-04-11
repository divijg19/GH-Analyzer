import type { Report } from "../api/client";

type TopReposProps = {
	repos: Report["top_repos"];
};

export default function TopRepos(props: TopReposProps) {
	return (
		<section class="card-section">
			<h3 class="section-title">Top Repositories</h3>
			{props.repos.length > 0 ? (
				<ul class="section-list">
					{props.repos.map((repo) => (
						<li>
							<span>{repo.name}</span>
							<span class="muted"> — {repo.size}</span>
						</li>
					))}
				</ul>
			) : (
				<p class="section-text">No repositories found.</p>
			)}
		</section>
	);
}
