export type SearchResult = {
	username: string;
	score: number;
	confidence: "high" | "moderate" | "low" | string;
	signals: {
		consistency: number;
		ownership: number;
		depth: number;
		activity: number;
	};
	reasons: string[];
};

export type SearchResponse = {
	query: string;
	mode: "dataset" | "live" | string;
	total: number;
	results: SearchResult[];
};

const SEARCH_URL = "http://localhost:8080/search";

export async function search(
	query: string,
	live: boolean,
): Promise<SearchResponse> {
	const response = await fetch(
		`${SEARCH_URL}?q=${encodeURIComponent(query)}&live=${live}`,
	);

	if (!response.ok) {
		throw new Error("request failed");
	}

	return (await response.json()) as SearchResponse;
}
