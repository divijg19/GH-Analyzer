import type { Report } from "../api/client";

type TopReposProps = {
	repos: Report["top_repos"];
};

export default function TopRepos(props: TopReposProps) {
	return (
		<section class="mt-4">
			<h3 class="text-sm font-semibold">Top Repositories</h3>
			{props.repos.length > 0 ? (
				<ul class="mt-2 space-y-1">
					{props.repos.map((repo) => (
						<li>
							<span>{repo.name}</span>
							<span class="text-gray-600"> — {repo.size}</span>
						</li>
					))}
				</ul>
			) : (
				<p class="mt-2 text-gray-600">No repositories found.</p>
			)}
		</section>
	);
}
