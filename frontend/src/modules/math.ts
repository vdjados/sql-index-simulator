export function cosineSimilarity(vecA: number[], vecB: number[]): number {
  const n = Math.min(vecA.length, vecB.length)
  if (n === 0) return 0

  let dot = 0
  let normA = 0
  let normB = 0

  for (let i = 0; i < n; i++) {
    const a = vecA[i] ?? 0
    const b = vecB[i] ?? 0
    dot += a * b
    normA += a * a
    normB += b * b
  }

  if (normA === 0 || normB === 0) return 0
  return dot / (Math.sqrt(normA) * Math.sqrt(normB))
}

