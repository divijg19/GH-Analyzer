import type { Report } from "../api/client";

type TopReposProps = {
	repos: Report["top_repos"];
};

export default function TopRepos(props: TopReposProps) {
	return (
		<section>
			<h3 class="text-xs uppercase tracking-wide text-gray-400">
				Top Repositories
			</h3>
			{props.repos.length > 0 ? (
				<ul class="mt-3 space-y-1">
					{props.repos.map((repo) => (
						<li>
							<span>{repo.name}</span>
							<span class="text-gray-600"> — {repo.size}</span>
						</li>
					))}
				</ul>
			) : (
				<p class="mt-3 text-gray-600">No repositories found.</p>
			)}
		</section>
	);
}
