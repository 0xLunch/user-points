# Parallel Claim with Points (Proof of Concept)
This repository is a monorepo containing the code for a Parallel TCG claim system with points and secure user authentication.

## Tech Stack:

- Frontend: Next.js (React) with Docker
- User Management & JWT Auth: Golang Microservice (server/)
- Guardian & Claim Logic: Cloudflare Worker (server/)
- Smart Contract: Solidity (contracts/)

## Project Structure:

- `/app`: React frontend application for user interface and claim requests.
- `/contracts`: Solidity smart contract for managing Parallel TCG NFTs and claim logic.
- `/server`: Golang microservices for user management, JWT generation, and Cloudflare Worker logic (optional gRPC server and database interaction).

### Getting Started:

Refer to the individual folders (/app, /contracts, /server) for specific setup and running instructions.

This project utilizes various tools and cloud services (Docker, Golang, Cloudflare Workers, etc.). Ensure you have the necessary environment and configurations set up before running the application.

#### Note:

This is a Proof of Concept project and is not intended for production use. It demonstrates core functionalities and interactions between technologies.