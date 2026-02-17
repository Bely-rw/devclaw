package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newHealthCmd cria o comando `devclaw health` para verificação de saúde.
// Usado pelo Docker HEALTHCHECK e monitoramento.
func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Verifica o estado de saúde do serviço",
		Long:  `Retorna o status de saúde do DevClaw. Usado por Docker HEALTHCHECK e monitoramento.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// TODO: Implementar verificação real (checar canais, scheduler, memória).
			fmt.Println(`{"status":"ok","version":"dev"}`)
			return nil
		},
	}
}
