<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Development Roadmap

**Last Updated:** December 9, 2025
**Strategy:** Phased, demand-driven development

## 🎯 Development Philosophy

1. **Maintain v3 as production-ready** - It works, it's stable, keep it that way
2. **Grow v4 incrementally** - Add features as users request them
3. **Validate before building** - Don't over-engineer without real demand

---

## 📅 Phase 1: Stable Foundation (Current - Q1 2026)

### Goal: Keep v3 Production-Ready, Monitor Demand

#### v3 SDK Maintenance ✅
- **Status:** v3.65.0-1 (Python SDK v3.65.0 parity)
- **Quality:** Production-ready, comprehensive tests
- **Action:** Monitor & maintain

**Ongoing Activities:**
- ✅ Monitor Python SDK v3.x releases (automated via GitHub Actions)
- 🔄 Sync with new v3.x releases as they appear
- 🐛 Fix bugs reported by users
- 📝 Improve documentation based on feedback
- 💬 Respond to community issues/PRs

#### v4 SDK Status ✅
- **Status:** v4.2.0-1 (Close() methods implemented)
- **Quality:** Minimal but functional
- **Action:** Available for early adopters

**Current Capabilities:**
- ✅ All 7 services with basic CRUD operations
- ✅ Context-first API design
- ✅ Explicit scope requirements
- ✅ Resource cleanup (Close() methods)

**Known Limitations:**
- ⚠️ No connection pooling
- ⚠️ No advanced rate limiting
- ⚠️ No token storage/refresh
- ⚠️ Minimal test coverage

#### Upstream Tracking ✅
- **Automation:** GitHub Actions workflow
- **Monitored:** Python SDK v3.x and v4.x
- **Action:** Creates issues for new releases
- **File:** `.github/upstream-versions.json`

---

## 📅 Phase 2: Demand-Driven Growth (Q1-Q2 2026)

### Goal: Add v4 Features Based on Real User Needs

#### Quick Wins Available

##### 1. AddAppFlowUserScope (2-3 days)
**Python SDK v4.2.0 Feature**
- **Effort:** Low
- **Value:** Medium
- **Users:** Those using Timers + Flows
- **Status:** 📋 Documented, ready to implement

**Implementation:**
```go
// v4/pkg/services/timers/client.go
func (c *Client) AddAppFlowUserScope(ctx context.Context, flowID string) error {
    // Register required scope dependencies for flow timers
}
```

##### 2. Basic Token Storage (1 week)
**High Value for Production Apps**
- **Effort:** Medium
- **Value:** High
- **Users:** All production applications
- **Status:** 📋 Can port from v3

**Features:**
- Memory-based token storage
- File-based token storage
- Automatic persistence
- Token expiry detection

##### 3. Rate Limiting (1 week)
**Essential for Production**
- **Effort:** Medium
- **Value:** High
- **Users:** High-volume API users
- **Status:** 📋 Can port from v3

**Features:**
- 429 response detection
- Exponential backoff
- Configurable retry strategies
- Rate limit headers parsing

##### 4. GARE Auto-Retry (1 week)
**Python SDK v4.2.0 Feature**
- **Effort:** High
- **Value:** Medium
- **Users:** Complex auth scenarios
- **Status:** 📋 Documented, needs design

**Features:**
- Automatic GARE detection
- Consent flow handling
- Configurable (opt-in)
- Token refresh integration

#### Implementation Trigger

**Add a feature when:**
1. ✅ User requests it (GitHub issue)
2. ✅ Production use case emerges
3. ✅ Blocks adoption of v4

**Don't build until:**
- ❌ "It would be nice to have"
- ❌ "For completeness"
- ❌ "Python SDK has it"

---

## 📅 Phase 3: Full Production v4 (If Demand Warrants)

### Goal: Complete v4 Feature Parity with v3

**Trigger Conditions:**
- Multiple users requesting v4 features
- v4 adoption growing
- Clear production use cases
- Community momentum

**Timeline:** 8-12 weeks (when triggered)

#### Infrastructure Buildout
1. **Connection Pooling** (1-2 weeks)
2. **Circuit Breakers** (1 week)
3. **Token Management** (2 weeks)
4. **Comprehensive Testing** (2-3 weeks)
5. **Documentation** (2 weeks)
6. **Migration Tools** (1 week)

**See:** [PYTHON_SDK_V4.2.0_TRACKING.md](./PYTHON_SDK_V4.2.0_TRACKING.md) for details

---

## 📊 Current Status Dashboard

### v3 SDK
| Metric | Status | Action |
|--------|--------|--------|
| **Version** | v3.65.0-1 | ✅ Current |
| **Python Parity** | v3.65.0 | ✅ Latest |
| **Tests** | 32 packages passing | ✅ Stable |
| **Recommendation** | Production use | ✅ Ready |

### v4 SDK
| Metric | Status | Action |
|--------|--------|--------|
| **Version** | v4.2.0-1 | ✅ Current |
| **Python Parity** | v4.2.0 (partial) | 🔶 Close() only |
| **Tests** | Core tests only | ⚠️ Minimal |
| **Recommendation** | Early adopters | 🔶 Use with caution |

### Upstream Tracking
| Branch | Latest | Go SDK | Status |
|--------|--------|--------|--------|
| **v3** | v3.65.0 | v3.65.0-1 | ✅ Synced |
| **v4** | v4.2.0 | v4.2.0-1 | 🔶 Partial |

---

## 🎯 Immediate Next Steps

### This Week
1. ✅ Commit this roadmap
2. 📋 Create tracking issues for incremental v4 features
3. 📝 Add contributing guidelines for v4 features
4. 🔍 Audit v3 for any quick improvements

### This Month
1. 📊 Monitor v3 usage/feedback
2. 🐛 Fix any reported v3 issues
3. 📚 Improve v3 documentation
4. 👀 Watch for upstream releases

### This Quarter
1. 📈 Assess v4 adoption
2. 🔧 Implement requested v4 features
3. 🧪 Add v4 test coverage
4. 📖 Create v4 usage examples

---

## 📋 Tracking Issues

### Create Issues For:

#### Ready for Implementation (When Requested)
- [ ] `feat(v4): implement AddAppFlowUserScope for timers` (Quick win)
- [ ] `feat(v4): add basic token storage` (High value)
- [ ] `feat(v4): implement rate limiting` (High value)
- [ ] `feat(v4): add GARE auto-retry` (Medium value)

#### Future Consideration
- [ ] `feat(v4): connection pooling` (Phase 3)
- [ ] `feat(v4): circuit breakers` (Phase 3)
- [ ] `feat(v4): comprehensive testing` (Phase 3)

---

## 🤝 Community Engagement

### Encourage Feedback
- Create CONTRIBUTING.md for v4
- Add templates for feature requests
- Document v4 limitations clearly
- Ask users about their needs

### Respond To
- Bug reports (priority)
- Feature requests (evaluate)
- Documentation improvements (quick wins)
- Examples/use cases (high value)

---

## 📈 Success Metrics

### v3 SDK
- ✅ Zero critical bugs
- ✅ Staying current with Python SDK
- ✅ Good documentation
- ✅ Positive user feedback

### v4 SDK
- 📊 Track adoption rate
- 📊 Monitor feature requests
- 📊 Measure user satisfaction
- 📊 Assess production usage

### Decision Points
- **If v4 adoption is low:** Keep minimal, focus on v3
- **If v4 adoption grows:** Invest in Phase 3
- **If features requested:** Implement incrementally

---

## 🔄 Review Schedule

- **Weekly:** Check for upstream releases
- **Monthly:** Review v3 health, v4 feedback
- **Quarterly:** Assess roadmap, adjust strategy
- **Annually:** Major version planning

---

## 📚 Related Documents

- [PYTHON_SDK_V4.2.0_TRACKING.md](./PYTHON_SDK_V4.2.0_TRACKING.md) - v4.2.0 implementation details
- [v4/RELEASE_NOTES_v4.2.0-1.md](./v4/RELEASE_NOTES_v4.2.0-1.md) - v4.2.0-1 release info
- [CLAUDE.md](./CLAUDE.md) - Development guidelines
- [ALIGNMENT.md](./ALIGNMENT.md) - Python SDK alignment strategy

---

**Strategy Summary:** Maintain excellent v3, grow v4 based on real demand, avoid premature optimization.

**Current Focus:** v3 stability + monitoring for upstream changes + responding to user needs.
