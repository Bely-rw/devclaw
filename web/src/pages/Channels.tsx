import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, Radio, QrCode, Wifi, WifiOff, Smartphone } from 'lucide-react'
import { api, type ChannelHealth } from '@/lib/api'
import { timeAgo } from '@/lib/utils'

/**
 * Página de gerenciamento de canais.
 * Mostra status de todos os canais configurados e permite
 * conectar/reconectar WhatsApp via QR code.
 */
export function Channels() {
  const navigate = useNavigate()
  const [channels, setChannels] = useState<ChannelHealth[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.channels.list()
      .then(setChannels)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const whatsapp = channels.find((ch) => ch.name === 'whatsapp')
  const otherChannels = channels.filter((ch) => ch.name !== 'whatsapp')

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center bg-[var(--color-dc-darker)]">
        <div className="h-10 w-10 rounded-full border-4 border-orange-500/30 border-t-orange-500 animate-spin" />
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto bg-[var(--color-dc-darker)]">
      <div className="mx-auto max-w-5xl px-8 py-10">
        <p className="text-[11px] font-bold uppercase tracking-[0.15em] text-gray-600">Comunicação</p>
        <h1 className="mt-1 text-2xl font-black text-white tracking-tight">Canais</h1>
        <p className="mt-2 text-base text-gray-500">Status e configuração dos canais de mensagem</p>

        <div className="mt-8 space-y-4">
          {/* WhatsApp — card dedicado com QR code */}
          {whatsapp && (
            <div
              className={`relative overflow-hidden rounded-2xl border p-6 transition-all ${
                whatsapp.connected
                  ? 'border-emerald-500/25 bg-emerald-500/[0.03]'
                  : 'border-orange-500/20 bg-orange-500/[0.02]'
              }`}
            >
              {whatsapp.connected && (
                <div className="absolute right-5 top-5">
                  <span className="rounded-full bg-emerald-500 px-3 py-1 text-[10px] font-bold text-white shadow-lg shadow-emerald-500/30">online</span>
                </div>
              )}

              <div className="flex items-start gap-5">
                <div className={`flex h-14 w-14 shrink-0 items-center justify-center rounded-xl ${
                  whatsapp.connected
                    ? 'bg-emerald-500/15 text-emerald-400'
                    : 'bg-orange-500/10 text-orange-400'
                }`}>
                  <svg viewBox="0 0 24 24" className="h-7 w-7" fill="currentColor">
                    <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                  </svg>
                </div>

                <div className="flex-1">
                  <h3 className="text-xl font-bold text-white">WhatsApp</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    {whatsapp.connected
                      ? whatsapp.last_msg_at && whatsapp.last_msg_at !== '0001-01-01T00:00:00Z'
                        ? `Última mensagem: ${timeAgo(whatsapp.last_msg_at)}`
                        : 'Conectado — aguardando mensagens'
                      : 'Desconectado — escaneie o QR code para conectar'}
                  </p>

                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    {whatsapp.error_count > 0 && (
                      <span className="flex items-center gap-1.5 rounded-full bg-amber-500/15 px-3 py-1 text-xs font-bold text-amber-400">
                        <AlertTriangle className="h-3 w-3" />
                        {whatsapp.error_count} erros
                      </span>
                    )}

                    <button
                      onClick={() => navigate('/channels/whatsapp')}
                      className={`flex cursor-pointer items-center gap-2 rounded-full px-4 py-1.5 text-xs font-bold text-white shadow-lg transition-all ${
                        whatsapp.connected
                          ? 'bg-zinc-700 shadow-none hover:bg-zinc-600'
                          : 'bg-gradient-to-r from-orange-500 to-amber-500 shadow-orange-500/20 hover:shadow-orange-500/30'
                      }`}
                    >
                      {whatsapp.connected ? (
                        <>
                          <Smartphone className="h-3.5 w-3.5" />
                          Gerenciar Conexão
                        </>
                      ) : (
                        <>
                          <QrCode className="h-3.5 w-3.5" />
                          Conectar via QR Code
                        </>
                      )}
                    </button>

                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                      whatsapp.connected
                        ? 'bg-emerald-500/10 text-emerald-400'
                        : 'bg-red-500/10 text-red-400'
                    }`}>
                      {whatsapp.connected ? 'Conectado' : 'Offline'}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Outros canais */}
          {otherChannels.map((ch) => (
            <div
              key={ch.name}
              className={`relative overflow-hidden rounded-2xl border p-6 transition-all ${
                ch.connected
                  ? 'border-emerald-500/25 bg-emerald-500/[0.03]'
                  : 'border-white/[0.06] bg-[var(--color-dc-dark)]'
              }`}
            >
              {ch.connected && (
                <div className="absolute right-5 top-5">
                  <span className="rounded-full bg-emerald-500 px-3 py-1 text-[10px] font-bold text-white shadow-lg shadow-emerald-500/30">online</span>
                </div>
              )}

              <div className="flex items-start gap-5">
                <div className={`flex h-14 w-14 shrink-0 items-center justify-center rounded-xl ${
                  ch.connected
                    ? 'bg-emerald-500/15 text-emerald-400'
                    : 'bg-white/[0.05] text-gray-500'
                }`}>
                  {ch.connected ? <Wifi className="h-7 w-7" /> : <WifiOff className="h-7 w-7" />}
                </div>

                <div className="flex-1">
                  <h3 className="text-xl font-bold capitalize text-white">{ch.name}</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    {ch.last_msg_at && ch.last_msg_at !== '0001-01-01T00:00:00Z'
                      ? `Última mensagem: ${timeAgo(ch.last_msg_at)}`
                      : 'Sem mensagens recentes'}
                  </p>

                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    {ch.error_count > 0 && (
                      <span className="flex items-center gap-1.5 rounded-full bg-amber-500/15 px-3 py-1 text-xs font-bold text-amber-400">
                        <AlertTriangle className="h-3 w-3" />
                        {ch.error_count} erros
                      </span>
                    )}
                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                      ch.connected
                        ? 'bg-emerald-500/10 text-emerald-400'
                        : 'bg-red-500/10 text-red-400'
                    }`}>
                      {ch.connected ? 'Conectado' : 'Offline'}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          ))}

          {/* Nenhum canal */}
          {channels.length === 0 && (
            <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-white/[0.08] py-20">
              <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-white/[0.04]">
                <Radio className="h-8 w-8 text-gray-700" />
              </div>
              <p className="mt-4 text-lg font-semibold text-gray-500">Nenhum canal configurado</p>
              <p className="mt-1 text-sm text-gray-600">Configure canais no config.yaml ou via setup wizard</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
