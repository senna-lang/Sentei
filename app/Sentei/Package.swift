// swift-tools-version: 5.10
import PackageDescription

let package = Package(
    name: "Sentei",
    platforms: [
        .macOS(.v14)
    ],
    targets: [
        .executableTarget(
            name: "Sentei",
            path: "Sources"
        ),
        .testTarget(
            name: "SenteiTests",
            dependencies: ["Sentei"],
            path: "Tests"
        ),
    ]
)
