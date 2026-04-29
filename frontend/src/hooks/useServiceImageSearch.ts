import { useEffect, useMemo, useRef, useState } from 'react'
import { cosineSimilarity } from '../modules/math'
import type { ApiService } from '../api/client'

export interface RankedService extends ApiService {
  score: number
  isVisible: boolean
  embedding?: number[]
}

export function useServiceImageSearch(inputItems: ApiService[], opts?: { threshold?: number; topK?: number }) {
  const threshold = Math.min(0.9, Math.max(0.4, opts?.threshold ?? 0.62))
  const topK = Math.min(50, Math.max(1, opts?.topK ?? 8))

  const [ready, setReady] = useState(false)
  const [progress, setProgress] = useState(0)
  const [imageEmbedding, setImageEmbedding] = useState<number[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [textEmbeddings, setTextEmbeddings] = useState<Record<string, number[]> | null>(null)

  const workerRef = useRef<Worker | null>(null)

  // (Re)load model + compute text embeddings whenever the item list changes.
  useEffect(() => {
    setError(null)
    setReady(false)
    setProgress(0)
    setTextEmbeddings(null)

    const w = new Worker(new URL('../workers/search.worker.ts', import.meta.url), { type: 'module' })
    workerRef.current = w

    w.onmessage = (e: MessageEvent) => {
      const { type, data } = e.data as { type: string; data: unknown }
      if (type === 'progress') {
        const d = data as { status?: string; progress?: number }
        if (d?.status === 'progress' && typeof d.progress === 'number') {
          setProgress(d.progress)
        }
        if (d?.status === 'ready') setReady(true)
      }
      if (type === 'text_embeddings_ready') {
        setTextEmbeddings((data as Record<string, number[]>) ?? {})
        setReady(true)
      }
      if (type === 'image_embedding_ready') {
        setImageEmbedding((data as number[]) ?? null)
      }
      if (type === 'error') {
        setError(typeof data === 'string' ? data : 'worker error')
      }
    }

    const initPayload = inputItems.map((i) => ({
      id: i.id,
      description: i.short_description_en?.trim() ? i.short_description_en : i.description,
    }))
    w.postMessage({ type: 'init', data: initPayload })

    return () => {
      w.terminate()
      if (workerRef.current === w) workerRef.current = null
    }
  }, [inputItems])

  const ranked: RankedService[] = useMemo(() => {
    const base: RankedService[] = inputItems.map((i) => ({ ...i, score: 0, isVisible: true }))
    if (!imageEmbedding || !textEmbeddings) return base

    const scored = base.map((i) => {
      const emb = textEmbeddings[i.id]
      if (!emb) return i
      const score = cosineSimilarity(imageEmbedding, emb)
      return { ...i, embedding: emb, score, isVisible: score >= threshold }
    })

    scored.sort((a, b) => b.score - a.score)
    const visible = scored.filter((x) => x.isVisible)

    // If nothing passes the threshold, still show TopK best matches so the UI
    // doesn't end up empty on "template" images.
    const candidates = visible.length > 0 ? visible : scored
    const limitedIds = new Set(candidates.slice(0, topK).map((x) => x.id))

    return scored.map((x) => ({
      ...x,
      isVisible: limitedIds.has(x.id) && (visible.length === 0 || x.isVisible),
    }))
  }, [imageEmbedding, inputItems, textEmbeddings, threshold, topK])

  const searchByImage = (file: File) => {
    setError(null)
    workerRef.current?.postMessage({ type: 'image', data: file })
  }

  const resetSearch = () => {
    setImageEmbedding(null)
  }

  return { items: ranked, ready, progress, imageEmbedding, threshold, topK, searchByImage, resetSearch, error }
}

