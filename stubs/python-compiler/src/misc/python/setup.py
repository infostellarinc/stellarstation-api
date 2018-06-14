import os
import setuptools

with open('README.md', 'r') as fh:
    long_description = fh.read()

setuptools.setup(
    name='stellarstation',
    version='0.0.4',
    author='StellarStation API Team',
    author_email='stellarstation-api-team@istellar.jp',
    description='Client stubs for accessing the StellarStation API',
    long_description=long_description,
    long_description_content_type='text/markdown',
    url='https://github.com/infostellarinc/stellarstation-api',
    packages=setuptools.find_packages(),
    classifiers=(
        'Programming Language :: Python :: 2',
        'Programming Language :: Python :: 3',
        'License :: OSI Approved :: Apache Software License',
        'Operating System :: OS Independent ',
    ),
    install_requires=(
        'grpcio',
        'protobuf',
    ),
)
