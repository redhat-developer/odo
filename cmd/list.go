package cmd

// 'ocdev list' is just an alias for 'ocdev component list'
func init() {
	rootCmd.AddCommand(componentListCmd)
}
