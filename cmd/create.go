package cmd

// 'ocdev crate' is just an alias for 'ocdev component create'
func init() {
	rootCmd.AddCommand(componentCreateCmd)
}
