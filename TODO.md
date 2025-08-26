# OttoMap Roadmap

## Current Status (v0.62.2)
✅ **Foundation Complete**
- [x] Reports parsed with Pigeon and rendered to WXX files

## MVP Release (v0.63.0) - Reverse Walker

### Core Features
- [ ] **Parser** - parse entire report
- [ ] **Walker** - walk backwards through the input and return MapFeatures_t by element
- [ ] **Render** - update to use new MapFeatures_t structure

### Technical Improvements
- [ ] **Parser** - better error handling
- [ ] **Mapping** - improve location tracking
- [ ] **Rendering** - handle terrain and tile conflicts better
- [ ] **Logging improvements** - capture logs to files
- [ ] **Cache optimization** - reparse reports on change instead of every run

### User Experience
- [ ] **Error Reporting** - provide consolidated error reports

### Documentation
- [ ] **User guide** - Complete setup, customization, and operation manual
- [ ] **Troubleshooting guide** - Common issues and solutions

### Deployment Features
- [ ] **Log rotation support** - Proper log file handling
