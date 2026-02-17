import { useParams } from 'react-router-dom'
import { useEffect, useRef } from 'react'
import {
  Terminal,
  GitBranch,
  Database,
  Globe,
  FileCode,
  Server,
  Wrench,
  Zap,
} from 'lucide-react'
import { ChatMessage } from '@/components/ChatMessage'
import { ChatInput } from '@/components/ChatInput'
import { useChat } from '@/hooks/useChat'

const SUGGESTIONS = [
  { icon: GitBranch, label: 'Git status', prompt: 'Qual o status do meu repositório git?' },
  { icon: Server, label: 'Processos', prompt: 'Liste os processos rodando na porta 3000' },
  { icon: Database, label: 'DB schema', prompt: 'Mostre o schema do banco de dados' },
  { icon: FileCode, label: 'Analisar código', prompt: 'Analise a estrutura do projeto atual' },
  { icon: Globe, label: 'API test', prompt: 'Faça um GET em https://httpbin.org/get' },
  { icon: Wrench, label: 'Docker ps', prompt: 'Liste os containers Docker rodando' },
]

export function Chat() {
  const { sessionId } = useParams<{ sessionId: string }>()
  const resolvedId = sessionId ? decodeURIComponent(sessionId) : 'webui:default'
  const { messages, streamingContent, isStreaming, error, sendMessage, abort } = useChat(resolvedId)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingContent])

  const hasMessages = messages.length > 0 || streamingContent

  return (
    <div className="flex flex-1 flex-col overflow-hidden bg-dc-darker">
      <div className="flex-1 overflow-y-auto">
        {!hasMessages ? (
          <div className="flex h-full flex-col items-center justify-center px-6">
            <div className="flex flex-col items-center -mt-12">
              {/* Logo */}
              <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-linear-to-br from-orange-500/20 to-amber-500/10 ring-1 ring-orange-500/20">
                <Terminal className="h-8 w-8 text-orange-400" />
              </div>

              <h2 className="mt-5 text-xl font-bold text-white">O que vamos fazer?</h2>
              <p className="mt-1.5 text-sm text-zinc-500">Pergunte qualquer coisa ou escolha uma sugestão</p>

              {/* Suggestions grid */}
              <div className="mt-8 grid w-full max-w-lg grid-cols-2 gap-2 sm:grid-cols-3">
                {SUGGESTIONS.map((s) => (
                  <button
                    key={s.label}
                    onClick={() => sendMessage(s.prompt)}
                    className="group flex cursor-pointer items-center gap-2.5 rounded-xl bg-zinc-800/40 px-3.5 py-3 text-left ring-1 ring-zinc-700/20 transition-all hover:bg-zinc-800/60 hover:ring-orange-500/20"
                  >
                    <s.icon className="h-4 w-4 shrink-0 text-zinc-500 transition-colors group-hover:text-orange-400" />
                    <span className="text-xs font-medium text-zinc-400 transition-colors group-hover:text-zinc-200">{s.label}</span>
                  </button>
                ))}
              </div>

              {/* Quick tips */}
              <div className="mt-6 flex items-center gap-4 text-[11px] text-zinc-600">
                <span className="flex items-center gap-1.5">
                  <Zap className="h-3 w-3 text-orange-500/50" />
                  70+ ferramentas nativas
                </span>
                <span className="h-3 w-px bg-zinc-700/50" />
                <span>Enter para enviar, Shift+Enter para nova linha</span>
              </div>
            </div>
          </div>
        ) : (
          <div className="mx-auto max-w-3xl space-y-1 px-6 py-8">
            {messages.map((msg, i) => (
              <ChatMessage key={i} role={msg.role} content={msg.content} toolName={msg.tool_name} toolInput={msg.tool_input} />
            ))}
            {streamingContent && (
              <ChatMessage role="assistant" content={streamingContent} isStreaming />
            )}
            {error && (
              <div className="rounded-xl border border-red-500/20 bg-red-500/5 px-5 py-4 text-sm text-red-400">{error}</div>
            )}
            <div ref={bottomRef} />
          </div>
        )}
      </div>

      <div className="mx-auto w-full max-w-3xl">
        <ChatInput onSend={sendMessage} onAbort={abort} isStreaming={isStreaming} />
      </div>
    </div>
  )
}
