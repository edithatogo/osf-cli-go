// Package zenodooai implements bounded, read-only Zenodo OAI-PMH harvesting.
// It is intentionally separate from the Zenodo REST adapter because protocol
// errors, metadata schemas, sets, and resumption tokens have distinct semantics.
package zenodooai
