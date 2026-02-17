import { useEffect, useState } from 'react'
import {
  Globe,
  Server,
  Shield,
  Save,
  Loader2,
  CheckCircle2,
  XCircle,
  ExternalLink,
  Network,
  Eye,
  EyeOff,
  Info,
  Plus,
  X,
} from 'lucide-react'
import { api } from '@/lib/api'
import type { DomainConfig } from '@/lib/api'

const inputClass =
  'flex h-11 w-full rounded-xl border border-zinc-700/50 bg-zinc-800/50 px-4 text-sm text-white placeholder:text-zinc-600 outline-none transition-all focus:border-orange-500/50 focus:ring-2 focus:ring-orange-500/10'

/**
 * Página de configuração de domínio e rede.
 * Permite configurar WebUI, Gateway API, Tailscale e CORS.
 */
export function Domain() {
  const [config, setConfig] = useState<DomainConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  /* Form state */
  const [webuiAddress, setWebuiAddress] = useState('')
  const [webuiToken, setWebuiToken] = useState('')
  const [showWebuiToken, setShowWebuiToken] = useState(false)

  const [gatewayEnabled, setGatewayEnabled] = useState(false)
  const [gatewayAddress, setGatewayAddress] = useState('')
  const [gatewayToken, setGatewayToken] = useState('')
  const [showGatewayToken, setShowGatewayToken] = useState(false)
  const [corsOrigins, setCorsOrigins] = useState<string[]>([])
  const [newCors, setNewCors] = useState('')

  const [tailscaleEnabled, setTailscaleEnabled] = useState(false)
  const [tailscaleServe, setTailscaleServe] = useState(false)
  const [tailscaleFunnel, setTailscaleFunnel] = useState(false)
  const [tailscalePort, setTailscalePort] = useState(8080)

  useEffect(() => {
    api.domain.get()
      .then((data) => {
        setConfig(data)
        setWebuiAddress(data.webui_address || ':8090')
        setGatewayEnabled(data.gateway_enabled)
        setGatewayAddress(data.gateway_address || ':8080')
        setCorsOrigins(data.cors_origins || [])
        setTailscaleEnabled(data.tailscale_enabled)
        setTailscaleServe(data.tailscale_serve)
        setTailscaleFunnel(data.tailscale_funnel)
        setTailscalePort(data.tailscale_port || 8080)
      })
      .catch(() => setMessage({ type: 'error', text: 'Erro ao carregar configuração' }))
      .finally(() => setLoading(false))
  }, [])

  const handleSave = async () => {
    setSaving(true)
    setMessage(null)
    try {
      await api.domain.update({
        webui_address: webuiAddress,
        webui_auth_token: webuiToken || undefined,
        gateway_enabled: gatewayEnabled,
        gateway_address: gatewayAddress,
        gateway_auth_token: gatewayToken || undefined,
        cors_origins: corsOrigins,
        tailscale_enabled: tailscaleEnabled,
        tailscale_serve: tailscaleServe,
        tailscale_funnel: tailscaleFunnel,
        tailscale_port: tailscalePort,
      })
      setMessage({ type: 'success', text: 'Configuração salva. Reinicie para aplicar alterações de porta.' })
      setWebuiToken('')
      setGatewayToken('')
    } catch {
      setMessage({ type: 'error', text: 'Erro ao salvar configuração' })
    } finally {
      setSaving(false)
    }
  }

  const addCorsOrigin = () => {
    const trimmed = newCors.trim()
    if (trimmed && !corsOrigins.includes(trimmed)) {
      setCorsOrigins([...corsOrigins, trimmed])
      setNewCors('')
    }
  }

  const removeCorsOrigin = (origin: string) => {
    setCorsOrigins(corsOrigins.filter((o) => o !== origin))
  }

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center bg-[var(--color-dc-darker)]">
        <div className="h-10 w-10 rounded-full border-4 border-orange-500/30 border-t-orange-500 animate-spin" />
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col overflow-hidden bg-[var(--color-dc-darker)]">
      <div className="mx-auto w-full max-w-3xl flex-1 overflow-y-auto px-8 py-10">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div>
            <p className="text-[11px] font-bold uppercase tracking-[0.15em] text-gray-600">Rede</p>
            <h1 className="mt-1 text-2xl font-black text-white tracking-tight">Domínio & Acesso</h1>
            <p className="mt-2 text-base text-gray-500">Configure como o DevClaw é acessado</p>
          </div>
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex cursor-pointer items-center gap-2 rounded-xl bg-gradient-to-r from-orange-500 to-amber-500 px-5 py-3 text-sm font-bold text-white shadow-lg shadow-orange-500/20 transition-all hover:shadow-orange-500/30 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            {saving ? 'Salvando...' : 'Salvar'}
          </button>
        </div>

        {/* Message */}
        {message && (
          <div className={`mt-6 rounded-2xl px-5 py-4 text-sm ring-1 ${
            message.type === 'success'
              ? 'bg-emerald-500/5 text-emerald-400 ring-emerald-500/20'
              : 'bg-red-500/5 text-red-400 ring-red-500/20'
          }`}>
            {message.text}
          </div>
        )}

        {/* Status bar */}
        {config && (
          <div className="mt-6 grid grid-cols-3 gap-3">
            <StatusCard
              label="WebUI"
              url={config.webui_url}
              active={true}
              secure={config.webui_auth_configured}
            />
            <StatusCard
              label="Gateway API"
              url={config.gateway_url}
              active={config.gateway_enabled}
              secure={config.gateway_auth_configured}
            />
            <StatusCard
              label="Tailscale"
              url={config.public_url || config.tailscale_url}
              active={config.tailscale_enabled}
              secure={true}
            />
          </div>
        )}

        {/* ── WebUI Section ── */}
        <Section icon={Globe} title="Web UI" description="Painel de controle e chat">
          <Field label="Endereço de escuta" hint="Host:porta onde o painel será servido">
            <input
              value={webuiAddress}
              onChange={(e) => setWebuiAddress(e.target.value)}
              placeholder=":8090"
              className={inputClass}
            />
          </Field>

          <Field label="Senha do painel" hint="Deixe em branco para manter a senha atual">
            <div className="relative">
              <input
                type={showWebuiToken ? 'text' : 'password'}
                value={webuiToken}
                onChange={(e) => setWebuiToken(e.target.value)}
                placeholder={config?.webui_auth_configured ? '••••••••' : 'Sem senha configurada'}
                className={`${inputClass} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowWebuiToken(!showWebuiToken)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors"
              >
                {showWebuiToken ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
          </Field>
        </Section>

        {/* ── Gateway API Section ── */}
        <Section icon={Server} title="Gateway API" description="API REST compatível com OpenAI + WebSocket">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-sm font-medium text-zinc-300">Ativar Gateway</span>
              <p className="text-xs text-zinc-500">Expõe a API HTTP em uma porta separada</p>
            </div>
            <Toggle value={gatewayEnabled} onChange={setGatewayEnabled} />
          </div>

          {gatewayEnabled && (
            <>
              <Field label="Endereço de escuta" hint="Porta da API (separada do WebUI)">
                <input
                  value={gatewayAddress}
                  onChange={(e) => setGatewayAddress(e.target.value)}
                  placeholder=":8080"
                  className={inputClass}
                />
              </Field>

              <Field label="Auth Token" hint="Bearer token para autenticação. Deixe em branco para manter.">
                <div className="relative">
                  <input
                    type={showGatewayToken ? 'text' : 'password'}
                    value={gatewayToken}
                    onChange={(e) => setGatewayToken(e.target.value)}
                    placeholder={config?.gateway_auth_configured ? '••••••••' : 'Sem token'}
                    className={`${inputClass} pr-10`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowGatewayToken(!showGatewayToken)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors"
                  >
                    {showGatewayToken ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </Field>

              <Field label="CORS Origins" hint="Domínios autorizados para acessar a API">
                <div className="space-y-2">
                  {corsOrigins.map((origin) => (
                    <div key={origin} className="flex items-center gap-2">
                      <span className="flex-1 rounded-lg bg-zinc-800/50 px-3 py-2 text-sm font-mono text-zinc-300">
                        {origin}
                      </span>
                      <button
                        onClick={() => removeCorsOrigin(origin)}
                        className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg text-zinc-500 transition-colors hover:bg-red-500/10 hover:text-red-400"
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  ))}
                  <div className="flex items-center gap-2">
                    <input
                      value={newCors}
                      onChange={(e) => setNewCors(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && addCorsOrigin()}
                      placeholder="https://example.com"
                      className={`${inputClass} flex-1`}
                    />
                    <button
                      onClick={addCorsOrigin}
                      disabled={!newCors.trim()}
                      className="flex h-11 w-11 cursor-pointer items-center justify-center rounded-xl border border-zinc-700/50 bg-zinc-800/50 text-zinc-400 transition-all hover:border-zinc-600 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
                    >
                      <Plus className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </Field>
            </>
          )}
        </Section>

        {/* ── Tailscale Section ── */}
        <Section icon={Network} title="Tailscale" description="Acesso remoto seguro com HTTPS automático">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-sm font-medium text-zinc-300">Ativar Tailscale</span>
              <p className="text-xs text-zinc-500">Requer Tailscale instalado e conectado</p>
            </div>
            <Toggle value={tailscaleEnabled} onChange={setTailscaleEnabled} />
          </div>

          {tailscaleEnabled && (
            <>
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm font-medium text-zinc-300">Tailscale Serve</span>
                  <p className="text-xs text-zinc-500">HTTPS acessível dentro da sua Tailnet</p>
                </div>
                <Toggle value={tailscaleServe} onChange={setTailscaleServe} />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <span className="text-sm font-medium text-zinc-300">Tailscale Funnel</span>
                  <p className="text-xs text-zinc-500">HTTPS acessível publicamente na internet</p>
                </div>
                <Toggle value={tailscaleFunnel} onChange={setTailscaleFunnel} />
              </div>

              <Field label="Porta local" hint="Porta que o Tailscale vai encaminhar (default: 8080)">
                <input
                  type="number"
                  value={tailscalePort}
                  onChange={(e) => setTailscalePort(parseInt(e.target.value) || 8080)}
                  placeholder="8080"
                  className={inputClass}
                />
              </Field>

              {config?.tailscale_hostname && (
                <div className="flex items-start gap-3 rounded-xl bg-zinc-800/30 px-4 py-3 ring-1 ring-zinc-700/30">
                  <Info className="mt-0.5 h-4 w-4 shrink-0 text-orange-400" />
                  <div>
                    <p className="text-sm text-zinc-300">
                      Hostname: <code className="text-orange-400">{config.tailscale_hostname}</code>
                    </p>
                    {config.tailscale_url && (
                      <a
                        href={config.tailscale_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="mt-1 flex items-center gap-1 text-xs text-orange-400/70 hover:text-orange-400 transition-colors"
                      >
                        <ExternalLink className="h-3 w-3" />
                        {config.tailscale_url}
                      </a>
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </Section>
      </div>
    </div>
  )
}

/* ── Reusable Components ── */

function Section({
  icon: Icon,
  title,
  description,
  children,
}: {
  icon: React.FC<{ className?: string }>
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <div className="mt-8">
      <div className="mb-5 flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-orange-500/10">
          <Icon className="h-4.5 w-4.5 text-orange-400" />
        </div>
        <div>
          <h2 className="text-base font-bold text-white">{title}</h2>
          <p className="text-xs text-zinc-500">{description}</p>
        </div>
      </div>
      <div className="space-y-5 rounded-2xl border border-white/[0.06] bg-[var(--color-dc-dark)]/80 p-6">
        {children}
      </div>
    </div>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div>
      <label className="mb-2 block text-sm font-medium text-zinc-300">{label}</label>
      {children}
      {hint && <p className="mt-1.5 text-xs text-zinc-500">{hint}</p>}
    </div>
  )
}

function Toggle({ value, onChange }: { value: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      onClick={() => onChange(!value)}
      className={`relative h-6 w-11 cursor-pointer rounded-full transition-colors ${
        value ? 'bg-orange-500' : 'bg-zinc-700'
      }`}
    >
      <span
        className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform ${
          value ? 'translate-x-5' : 'translate-x-0.5'
        }`}
      />
    </button>
  )
}

function StatusCard({
  label,
  url,
  active,
  secure,
}: {
  label: string
  url?: string
  active: boolean
  secure: boolean
}) {
  return (
    <div className="rounded-xl border border-white/[0.06] bg-[var(--color-dc-dark)]/80 px-4 py-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-zinc-400">{label}</span>
        <div className="flex items-center gap-1.5">
          {active ? (
            <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
          ) : (
            <XCircle className="h-3.5 w-3.5 text-zinc-600" />
          )}
          {active && secure && <Shield className="h-3 w-3 text-orange-400" />}
        </div>
      </div>
      {url ? (
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="mt-1.5 flex items-center gap-1 text-xs font-mono text-zinc-300 hover:text-orange-400 transition-colors truncate"
        >
          <ExternalLink className="h-3 w-3 shrink-0" />
          {url}
        </a>
      ) : (
        <p className="mt-1.5 text-xs text-zinc-600">{active ? 'Ativo' : 'Desativado'}</p>
      )}
    </div>
  )
}
