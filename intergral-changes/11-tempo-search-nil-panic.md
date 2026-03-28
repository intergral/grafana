# 11. Tempo Search Nil Panic Fix

> **Note:** Very likely fixed upstream. Check before reimplementing.

**Problem:** A nil pointer dereference when `model.TableType` is nil in Tempo search results.

**Solution:** Default `tableType` to the traces table type when nil before using it.