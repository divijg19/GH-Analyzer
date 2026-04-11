export type Report = {
	username: string;
	scores: {
		ownership: number;
		consistency: number;
		depth: number;
		overall: number;
	};
	summary: string;
	highlights: string[];
	top_repos: {
		name: string;
		size: number;
	}[];
};

type ErrorPayload = {
	error?: string;
};

const ANALYZE_URL = "http://localhost:8080/analyze";

export async function fetchReport(username: string): Promise<Report> {
	const response = await fetch(
		`${ANALYZE_URL}?username=${encodeURIComponent(username)}`,
	);

	let payload: Report | ErrorPayload | null = null;
	try {
		payload = (await response.json()) as Report | ErrorPayload;
	} catch {
		payload = null;
	}

	if (!response.ok) {
		const message =
			typeof payload === "object" &&
			payload !== null &&
			"error" in payload &&
			typeof payload.error === "string"
				? payload.error
				: "analysis request failed";
		throw new Error(message);
	}

	if (!payload) {
		throw new Error("invalid API response");
	}

	return payload as Report;
}
