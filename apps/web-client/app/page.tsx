import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import {
	Package,
	Truck,
	Factory,
	BarChart3,
	AlertTriangle,
	TrendingUp,
	Users,
	MapPin,
} from "lucide-react";

export default function Home() {
	return (
		<div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
			{/* Header */}
			<header className="border-b bg-white/80 backdrop-blur-sm dark:bg-slate-900/80">
				<div className="container mx-auto px-4 py-4">
					<div className="flex items-center justify-between">
						<div className="flex items-center space-x-2">
							<Package className="h-8 w-8 text-blue-600" />
							<h1 className="text-2xl font-bold text-slate-900 dark:text-white">
								SupplyChain Pro
							</h1>
						</div>
						<div className="flex items-center space-x-4">
							<Badge
								variant="outline"
								className="text-green-600 border-green-600"
							>
								All Systems Operational
							</Badge>
							<Button variant="outline">Sign In</Button>
						</div>
					</div>
				</div>
			</header>

			{/* Main Content */}
			<main className="container mx-auto px-4 py-8">
				{/* Hero Section */}
				<section className="text-center mb-12">
					<h2 className="text-4xl font-bold text-slate-900 dark:text-white mb-4">
						Real-Time Supply Chain Management
					</h2>
					<p className="text-xl text-slate-600 dark:text-slate-300 mb-8 max-w-3xl mx-auto">
						Monitor, optimize, and manage your entire supply chain in real-time
						with advanced analytics and automation
					</p>
					<div className="flex justify-center space-x-4">
						<Button size="lg" className="bg-blue-600 hover:bg-blue-700">
							Get Started
						</Button>
						<Button variant="outline" size="lg">
							View Demo
						</Button>
					</div>
				</section>

				{/* Dashboard Overview */}
				<section className="mb-12">
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
						<Card>
							<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
								<CardTitle className="text-sm font-medium">
									Total Orders
								</CardTitle>
								<Package className="h-4 w-4 text-muted-foreground" />
							</CardHeader>
							<CardContent>
								<div className="text-2xl font-bold">2,847</div>
								<p className="text-xs text-muted-foreground">
									<span className="text-green-600">+12%</span> from last month
								</p>
							</CardContent>
						</Card>

						<Card>
							<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
								<CardTitle className="text-sm font-medium">
									In Transit
								</CardTitle>
								<Truck className="h-4 w-4 text-muted-foreground" />
							</CardHeader>
							<CardContent>
								<div className="text-2xl font-bold">1,234</div>
								<p className="text-xs text-muted-foreground">
									<span className="text-blue-600">+5%</span> from last week
								</p>
							</CardContent>
						</Card>

						<Card>
							<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
								<CardTitle className="text-sm font-medium">
									Suppliers
								</CardTitle>
								<Factory className="h-4 w-4 text-muted-foreground" />
							</CardHeader>
							<CardContent>
								<div className="text-2xl font-bold">89</div>
								<p className="text-xs text-muted-foreground">
									<span className="text-green-600">+3</span> new this month
								</p>
							</CardContent>
						</Card>

						<Card>
							<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
								<CardTitle className="text-sm font-medium">
									Efficiency
								</CardTitle>
								<BarChart3 className="h-4 w-4 text-muted-foreground" />
							</CardHeader>
							<CardContent>
								<div className="text-2xl font-bold">94.2%</div>
								<p className="text-xs text-muted-foreground">
									<span className="text-green-600">+2.1%</span> improvement
								</p>
							</CardContent>
						</Card>
					</div>

					{/* Detailed Analytics */}
					<Tabs defaultValue="overview" className="w-full">
						<TabsList className="grid w-full grid-cols-4">
							<TabsTrigger value="overview">Overview</TabsTrigger>
							<TabsTrigger value="inventory">Inventory</TabsTrigger>
							<TabsTrigger value="logistics">Logistics</TabsTrigger>
							<TabsTrigger value="analytics">Analytics</TabsTrigger>
						</TabsList>

						<TabsContent value="overview" className="space-y-6">
							<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
								<Card>
									<CardHeader>
										<CardTitle className="flex items-center">
											<AlertTriangle className="mr-2 h-5 w-5 text-amber-500" />
											Recent Alerts
										</CardTitle>
									</CardHeader>
									<CardContent className="space-y-4">
										<div className="flex items-center justify-between">
											<div className="flex items-center space-x-2">
												<Badge variant="destructive">Critical</Badge>
												<span className="text-sm">
													Supplier XYZ delayed
												</span>
											</div>
											<span className="text-xs text-muted-foreground">
												2h ago
											</span>
										</div>
										<div className="flex items-center justify-between">
											<div className="flex items-center space-x-2">
												<Badge variant="secondary">Warning</Badge>
												<span className="text-sm">
													Low inventory: Widget A
												</span>
											</div>
											<span className="text-xs text-muted-foreground">
												4h ago
											</span>
										</div>
										<div className="flex items-center justify-between">
											<div className="flex items-center space-x-2">
												<Badge variant="outline">Info</Badge>
												<span className="text-sm">
													New route optimized
												</span>
											</div>
											<span className="text-xs text-muted-foreground">
												6h ago
											</span>
										</div>
									</CardContent>
								</Card>

								<Card>
									<CardHeader>
										<CardTitle className="flex items-center">
											<TrendingUp className="mr-2 h-5 w-5 text-green-500" />
											Performance Metrics
										</CardTitle>
									</CardHeader>
									<CardContent className="space-y-4">
										<div>
											<div className="flex justify-between text-sm mb-2">
												<span>Order Fulfillment Rate</span>
												<span>96%</span>
											</div>
											<Progress value={96} className="h-2" />
										</div>
										<div>
											<div className="flex justify-between text-sm mb-2">
												<span>On-Time Delivery</span>
												<span>89%</span>
											</div>
											<Progress value={89} className="h-2" />
										</div>
										<div>
											<div className="flex justify-between text-sm mb-2">
												<span>Inventory Accuracy</span>
												<span>98%</span>
											</div>
											<Progress value={98} className="h-2" />
										</div>
									</CardContent>
								</Card>
							</div>
						</TabsContent>

						<TabsContent value="inventory" className="space-y-6">
							<Card>
								<CardHeader>
									<CardTitle>Inventory Status</CardTitle>
									<CardDescription>
										Current stock levels across all warehouses
									</CardDescription>
								</CardHeader>
								<CardContent>
									<div className="space-y-4">
										<div className="grid grid-cols-3 gap-4">
											<div className="text-center">
												<div className="text-2xl font-bold text-green-600">
													342
												</div>
												<div className="text-sm text-muted-foreground">
													In Stock
												</div>
											</div>
											<div className="text-center">
												<div className="text-2xl font-bold text-amber-600">
													28
												</div>
												<div className="text-sm text-muted-foreground">
													Low Stock
												</div>
											</div>
											<div className="text-center">
												<div className="text-2xl font-bold text-red-600">
													7
												</div>
												<div className="text-sm text-muted-foreground">
													Out of Stock
												</div>
											</div>
										</div>
									</div>
								</CardContent>
							</Card>
						</TabsContent>

						<TabsContent value="logistics" className="space-y-6">
							<Card>
								<CardHeader>
									<CardTitle className="flex items-center">
										<MapPin className="mr-2 h-5 w-5" />
										Active Shipments
									</CardTitle>
								</CardHeader>
								<CardContent>
									<div className="space-y-3">
										<div className="flex items-center justify-between p-3 border rounded">
											<div className="flex items-center space-x-3">
												<Truck className="h-5 w-5 text-blue-600" />
												<div>
													<div className="font-medium">
														Shipment #SCM-2025-001
													</div>
													<div className="text-sm text-muted-foreground">
														New York → Los Angeles
													</div>
												</div>
											</div>
											<Badge
												variant="outline"
												className="text-green-600"
											>
												On Time
											</Badge>
										</div>
										<div className="flex items-center justify-between p-3 border rounded">
											<div className="flex items-center space-x-3">
												<Truck className="h-5 w-5 text-amber-600" />
												<div>
													<div className="font-medium">
														Shipment #SCM-2025-002
													</div>
													<div className="text-sm text-muted-foreground">
														Chicago → Miami
													</div>
												</div>
											</div>
											<Badge variant="secondary">Delayed</Badge>
										</div>
									</div>
								</CardContent>
							</Card>
						</TabsContent>

						<TabsContent value="analytics" className="space-y-6">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
								<Card>
									<CardHeader>
										<CardTitle>Cost Savings</CardTitle>
									</CardHeader>
									<CardContent>
										<div className="text-3xl font-bold text-green-600 mb-2">
											$2.4M
										</div>
										<p className="text-sm text-muted-foreground">
											Saved this quarter through optimization
										</p>
									</CardContent>
								</Card>
								<Card>
									<CardHeader>
										<CardTitle>Predictive Insights</CardTitle>
									</CardHeader>
									<CardContent>
										<div className="text-3xl font-bold text-blue-600 mb-2">
											94%
										</div>
										<p className="text-sm text-muted-foreground">
											Demand forecast accuracy
										</p>
									</CardContent>
								</Card>
							</div>
						</TabsContent>
					</Tabs>
				</section>

				{/* Quick Actions */}
				<section className="mb-12">
					<h3 className="text-2xl font-bold mb-6">Quick Actions</h3>
					<div className="grid grid-cols-1 md:grid-cols-3 gap-6">
						<Card className="cursor-pointer hover:shadow-lg transition-shadow">
							<CardHeader className="text-center">
								<Package className="mx-auto h-12 w-12 text-blue-600 mb-2" />
								<CardTitle>Create Order</CardTitle>
								<CardDescription>
									Start a new supply chain order
								</CardDescription>
							</CardHeader>
						</Card>
						<Card className="cursor-pointer hover:shadow-lg transition-shadow">
							<CardHeader className="text-center">
								<Users className="mx-auto h-12 w-12 text-green-600 mb-2" />
								<CardTitle>Manage Suppliers</CardTitle>
								<CardDescription>
									View and update supplier information
								</CardDescription>
							</CardHeader>
						</Card>
						<Card className="cursor-pointer hover:shadow-lg transition-shadow">
							<CardHeader className="text-center">
								<BarChart3 className="mx-auto h-12 w-12 text-purple-600 mb-2" />
								<CardTitle>Generate Reports</CardTitle>
								<CardDescription>
									Create detailed analytics reports
								</CardDescription>
							</CardHeader>
						</Card>
					</div>
				</section>
			</main>

			{/* Footer */}
			<footer className="bg-white dark:bg-slate-900 border-t">
				<div className="container mx-auto px-4 py-8">
					<div className="flex justify-between items-center">
						<p className="text-sm text-muted-foreground">
							© 2025 SupplyChain Pro. All rights reserved.
						</p>
						<div className="flex space-x-4">
							<Button variant="ghost" size="sm">
								Privacy
							</Button>
							<Button variant="ghost" size="sm">
								Terms
							</Button>
							<Button variant="ghost" size="sm">
								Support
							</Button>
						</div>
					</div>
				</div>
			</footer>
		</div>
	);
}
