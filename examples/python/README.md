# Installation Instructions for StellarStation Python Examples

1. Update your System
```bash
$ apt update && apt upgrade
```

2. Install Dependencies
Make sure you're using Python 3.10 or later.
```bash
$ apt install python3.10-venv
$ pip3 install pip --upgrade
```

3. Activate a new Virtual Environment
```bash
$ python3 -m venv venv
$ source venv/bin/activate
```

4. Install Required Packages for the Examples
```bash
$ pip install -r requirements.txt
```

5. Run one of the Examples
```bash
$  python3 one_of_the_examples.py
```
