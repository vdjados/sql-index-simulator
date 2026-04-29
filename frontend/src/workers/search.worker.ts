import {
  env,
  AutoTokenizer,
  AutoProcessor,
  SiglipTextModel,
  SiglipVisionModel,
  RawImage,
} from '@huggingface/transformers'

env.allowLocalModels = false
env.allowRemoteModels = true

const MODEL_ID = 'Xenova/siglip-base-patch16-224'
const EMBEDDING_SIZE = 768

class SiglipService {
  static tokenizer: Awaited<ReturnType<typeof AutoTokenizer.from_pretrained>> | null = null
  static processor: Awaited<ReturnType<typeof AutoProcessor.from_pretrained>> | null = null
  static textModel: Awaited<ReturnType<typeof SiglipTextModel.from_pretrained>> | null = null
  static visionModel: Awaited<ReturnType<typeof SiglipVisionModel.from_pretrained>> | null = null

  static async init(progress_callback?: (data: unknown) => void) {
    if (this.tokenizer) return
    const options = { device: 'wasm', dtype: 'q8' } as const

    this.tokenizer = await AutoTokenizer.from_pretrained(MODEL_ID, { progress_callback })
    this.processor = await AutoProcessor.from_pretrained(MODEL_ID, { progress_callback })
    this.textModel = await SiglipTextModel.from_pretrained(MODEL_ID, { ...options, progress_callback })
    this.visionModel = await SiglipVisionModel.from_pretrained(MODEL_ID, { ...options, progress_callback })
  }
}

type InitItem = { id: string; description: string }
type EmbeddingsMap = Record<string, number[]>

async function embedTexts(items: InitItem[]): Promise<EmbeddingsMap> {
  if (!SiglipService.tokenizer || !SiglipService.textModel) throw new Error('not ready')

  const descriptions = items.map((i) => i.description)
  const textInputs = await SiglipService.tokenizer(descriptions, {
    padding: 'max_length',
    truncation: true,
  })
  const { pooler_output } = await SiglipService.textModel(textInputs)
  const data = pooler_output.data as Float32Array

  const out: EmbeddingsMap = {}
  for (let i = 0; i < items.length; i++) {
    const item = items[i]
    if (!item) continue
    const start = i * EMBEDDING_SIZE
    const end = start + EMBEDDING_SIZE
    out[item.id] = Array.from(data.slice(start, end))
  }
  return out
}

async function embedImage(file: File): Promise<number[]> {
  if (!SiglipService.processor || !SiglipService.visionModel) throw new Error('not ready')
  const url = URL.createObjectURL(file)
  try {
    const img = await RawImage.read(url)
    const inputs = await SiglipService.processor(img)
    const { pooler_output } = await SiglipService.visionModel(inputs)
    return Array.from(pooler_output.data as Float32Array)
  } finally {
    URL.revokeObjectURL(url)
  }
}

self.addEventListener('message', async (event: MessageEvent) => {
  const { type, data } = event.data as { type: string; data: unknown }

  try {
    if (type === 'init') {
      await SiglipService.init((msg) => {
        ;(self as unknown as Worker).postMessage({ type: 'progress', data: msg })
      })
      const items = (Array.isArray(data) ? data : []) as InitItem[]
      const embeddings = await embedTexts(items)
      ;(self as unknown as Worker).postMessage({ type: 'text_embeddings_ready', data: embeddings })
      return
    }

    if (type === 'image') {
      const file = data as File
      const emb = await embedImage(file)
      ;(self as unknown as Worker).postMessage({ type: 'image_embedding_ready', data: emb })
      return
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    ;(self as unknown as Worker).postMessage({ type: 'error', data: msg })
  }
})

