// Package extsdk helps to write go-c8y-cli (c8y) extensions as compiled Go binaries.
//
// c8y runs extension binaries with a runtime context in the environment: the active
// session, the extension's folder, the path to the c8y executable, and global settings
// such as dry run. Use FromEnv to read it, Context.Client to create a Cumulocity client,
// and Print to write output which c8y can format (views, --select, --output etc.) for
// commands declared with "output: json" in extension.yaml.
//
//	func main() {
//		root := extsdk.NewRootCommand("c8y-hello")
//		root.AddCommand(&cobra.Command{
//			Use: "list",
//			RunE: func(cmd *cobra.Command, args []string) error {
//				ctx, err := extsdk.FromEnv()
//				if err != nil {
//					return err
//				}
//				client, err := ctx.Client()
//				if err != nil {
//					return err
//				}
//				col, _, err := client.Inventory.GetManagedObjects(cmd.Context(), nil)
//				if err != nil {
//					return extsdk.CheckError(err)
//				}
//				return extsdk.Print(col.ManagedObjects)
//			},
//		})
//		extsdk.Execute(root)
//	}
package extsdk
