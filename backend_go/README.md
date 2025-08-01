# Fantasy Esports Backend - Enterprise Grade GoLang API

🚀 **Complete Dream11-style Fantasy Esports Platform Backend**

## ⭐ Key Features Implemented

### 🔐 Authentication System
- **One-step OTP authentication** (Dream11 style)
- Mobile number registration/login
- JWT token management (access + refresh)
- Console OTP printing (for development)

### 🎮 Games & Tournament Management
- Dynamic game configuration with JSON scoring rules
- Tournament and match management
- Player and team databases
- Real-time match scheduling

### 🏆 Contest System
- Public and private contest creation
- Fantasy team composition with budget constraints
- Captain/Vice-captain multipliers (2x/1.5x)
- Real-time leaderboards

### 💰 Multi-Balance Wallet System
- Three balance types: Bonus, Deposit, Winning
- Payment gateway integration structure
- Withdrawal management
- Transaction history

### ⚡ **Manual Match Scoring System (Crown Jewel)**
- **Real-time admin scoring interface**
- **Live event addition** (kills, deaths, assists, objectives)
- **Bulk event updates**
- **WebSocket integration** for live updates
- **Fantasy point recalculation**
- **Dynamic leaderboard updates**

### 🛡️ Admin Dashboard
- Complete user management
- KYC verification system
- Financial transaction monitoring
- System configuration management
- **Live scoring dashboard**

### 🌐 Additional Features
- **Cloudinary CDN integration** for image uploads
- **Comprehensive API documentation** (Swagger)
- **PostgreSQL database** with 20+ tables
- **Referral system**
- **Role-based access control**

## 🏗️ Architecture

### Database Schema (20+ Tables)
- Users & Authentication
- Games & Tournaments
- Matches & Events
- Contests & Participants
- Fantasy Teams & Players
- Wallet & Transactions
- Admin & Configuration
- KYC Documents
- Referrals

### API Endpoints (50+ Endpoints)
```
/api/v1/auth/*          - Authentication
/api/v1/users/*         - User management
/api/v1/games/*         - Games & tournaments
/api/v1/matches/*       - Match information
/api/v1/contests/*      - Contest management
/api/v1/teams/*         - Fantasy team creation
/api/v1/leaderboards/*  - Leaderboards
/api/v1/wallet/*        - Wallet operations
/api/v1/admin/*         - Admin operations
/api/v1/admin/matches/* - Manual scoring (⭐ Crown Jewel)
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL database (provided)
- Cloudinary account (configured)

### Installation & Setup
```bash
# Clone and navigate
cd backend_go

# Install dependencies
go mod tidy

# Generate Swagger docs
swag init --dir cmd/server --output docs

# Build the server
go build -o fantasy-esports-server cmd/server/main.go

# Run the server
./fantasy-esports-server
```

### Environment Configuration
```env
DATABASE_URL=postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres
CLOUDINARY_URL=cloudinary://684824545515239:TaHGxQ0hRDOQW4mYN0hfHGPRxcc@dwnxysjxp
JWT_SECRET=your-super-secret-jwt-key-for-fantasy-esports-platform-2025
PORT=8080
```

## 📊 API Testing Examples

### 1. User Registration (One-step)
```bash
# Step 1: Verify Mobile
curl -X POST http://localhost:8080/api/v1/auth/verify-mobile \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "+919876543210",
    "country_code": "+91",
    "device_id": "test-device-123",
    "app_version": "1.0.0",
    "platform": "android"
  }'

# Step 2: Verify OTP (check console for OTP)
curl -X POST http://localhost:8080/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "SESSION_ID_FROM_STEP_1",
    "otp": "OTP_FROM_CONSOLE",
    "device_info": {
      "platform": "android",
      "device_id": "test-device-123",
      "app_version": "1.0.0"
    },
    "profile_data": {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "date_of_birth": "1995-06-15T00:00:00Z",
      "state": "Maharashtra"
    }
  }'
```

### 2. Admin Login & Manual Scoring
```bash
# Admin Login
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# Get Live Scoring Matches
curl -X GET http://localhost:8080/api/v1/admin/matches/live-scoring \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Add Match Event (Manual Scoring)
curl -X POST http://localhost:8080/api/v1/admin/matches/1/events \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "player_id": 101,
    "event_type": "kill",
    "points": 2.0,
    "round_number": 5,
    "timestamp": "2025-07-30T15:25:30Z",
    "description": "Entry frag on Haven A site"
  }'
```

### 3. Wallet Operations
```bash
# Get Balance
curl -X GET http://localhost:8080/api/v1/wallet/balance \
  -H "Authorization: Bearer USER_TOKEN"

# Deposit Money
curl -X POST http://localhost:8080/api/v1/wallet/deposit \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 500.00,
    "payment_method": "razorpay",
    "return_url": "https://app.example.com/payment-success"
  }'
```

## 🔧 Manual Scoring System (Crown Jewel Features)

### Real-time Match Scoring
- **Live event tracking**: Kills, deaths, assists, objectives
- **Bulk event updates** for efficiency
- **Fantasy point calculation** with multipliers
- **WebSocket connections** for real-time updates
- **Admin dashboard** with match statistics

### WebSocket Integration
```javascript
// Connect to live scoring WebSocket
const ws = new WebSocket('ws://localhost:8080/api/v1/admin/ws/live-scoring/123?token=ADMIN_TOKEN');

// Add event via WebSocket
ws.send(JSON.stringify({
  action: "add_event",
  data: {
    player_id: 101,
    event_type: "kill",
    points: 2.0,
    timestamp: "2025-07-30T15:25:30Z"
  }
}));
```

## 📚 Documentation

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **API Base**: http://localhost:8080/api/v1/

## 🎯 Production Features

### Scalability
- Clean architecture with separation of concerns
- Database connection pooling
- JWT-based stateless authentication
- CDN integration for media files

### Security
- Role-based access control
- JWT token expiration
- Input validation and sanitization
- SQL injection prevention

### Monitoring
- Comprehensive logging
- Request/response tracking
- Error handling and reporting
- Health check endpoints

## 🏆 Achievements

✅ **Complete PostgreSQL schema** (20+ tables)
✅ **50+ API endpoints** implemented
✅ **One-step authentication** (Dream11 style)
✅ **Manual scoring system** with WebSocket
✅ **Multi-balance wallet** system
✅ **Cloudinary CDN** integration
✅ **Admin dashboard** functionality
✅ **Swagger documentation** generated
✅ **Fantasy team management** with validation
✅ **Real-time leaderboards**
✅ **KYC verification** system

## 🚀 Next Steps (Production Ready)

1. **Integration Testing** - Comprehensive API testing
2. **Payment Gateway** - Razorpay/PayPal integration
3. **SMS/Email** - Notification system
4. **Push Notifications** - Real-time alerts
5. **Caching** - Redis integration
6. **Rate Limiting** - API protection
7. **Deployment** - Docker containerization

---

**Built with ❤️ using GoLang, PostgreSQL, and Enterprise-grade architecture principles.**